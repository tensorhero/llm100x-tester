package stages

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

const (
	// ServerStartupTimeout is the maximum time to wait for Flask server to start
	ServerStartupTimeout = 10 * time.Second

	// ConnectTimeout is the timeout for each connection attempt
	ConnectTimeout = 100 * time.Millisecond

	// CheckInterval is the interval between server readiness checks
	CheckInterval = 100 * time.Millisecond
)

func financeTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "finance",
		Timeout:       120 * time.Second,
		TestFunc:      testFinance,
		RequiredFiles: []string{"app.py"},
	}
}

// flaskServer manages a Flask application process
type flaskServer struct {
	cmd     *exec.Cmd
	port    int
	baseURL string
}

// startFlaskServer starts the Flask application and returns a flaskServer
func startFlaskServer(workDir string, port int, logger interface {
	Infof(format string, args ...interface{})
}) (*flaskServer, error) {
	// Find the venv Python - look for .venv in the project directory first
	venvPython := filepath.Join(workDir, ".venv", "bin", "python3")
	pythonPath := "python3" // fallback
	if _, err := os.Stat(venvPython); err == nil {
		pythonPath = venvPython
		logger.Infof("Using venv Python: %s", pythonPath)
	} else {
		logger.Infof("venv not found at %s, using system python3", venvPython)
	}

	// Set environment variables for Flask
	env := os.Environ()
	env = append(env, "FLASK_APP=app.py")
	env = append(env, "FLASK_ENV=development")
	env = append(env, "TENSORHERO_TEST_MODE=1") // Enable mock lookup
	env = append(env, fmt.Sprintf("FLASK_RUN_PORT=%d", port))

	// Start Flask using python -m flask run
	cmd := exec.Command(pythonPath, "-m", "flask", "run", "--port", fmt.Sprintf("%d", port))
	cmd.Dir = workDir
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Capture stdout/stderr for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Flask: %v", err)
	}

	server := &flaskServer{
		cmd:     cmd,
		port:    port,
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
	}

	// Wait for server to be ready
	if err := server.waitForReady(ServerStartupTimeout); err != nil {
		server.stop()
		return nil, fmt.Errorf("Flask server failed to start: %v\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	return server, nil
}

// waitForReady waits for the server to accept connections
func (s *flaskServer) waitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", s.port), ConnectTimeout)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(CheckInterval)
	}
	return fmt.Errorf("server did not become ready within %v", timeout)
}

// stop kills the Flask server process
func (s *flaskServer) stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		syscall.Kill(-s.cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(500 * time.Millisecond)
		syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)
	}
}

// httpClient wraps http.Client with session/cookie support
type httpClient struct {
	client  *http.Client
	baseURL string
}

// newHTTPClient creates a new HTTP client with cookie jar
func newHTTPClient(baseURL string) (*httpClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &httpClient{
		client: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
		baseURL: baseURL,
	}, nil
}

// get performs a GET request
func (c *httpClient) get(path string) (*http.Response, string, error) {
	resp, err := c.client.Get(c.baseURL + path)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp, string(body), err
}

// postForm performs a POST request with form data
func (c *httpClient) postForm(path string, data url.Values) (*http.Response, string, error) {
	resp, err := c.client.PostForm(c.baseURL+path, data)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp, string(body), err
}

func testFinance(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// Convert to absolute path
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	workDir = absWorkDir
	logger.Infof("Working directory: %s", workDir)

	// Copy finance.db to a temp location to avoid modifying the original
	origDB := filepath.Join(workDir, "finance.db")
	tempDir, err := os.MkdirTemp("", "finance_test_*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy all files to temp dir
	if err := copyDir(workDir, tempDir); err != nil {
		return fmt.Errorf("failed to copy files to temp dir: %v", err)
	}
	// Create fresh finance.db with transactions table
	tempDB := filepath.Join(tempDir, "finance.db")
	if err := resetDatabase(tempDB); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	_ = origDB // Keep compiler happy

	// Find an available port
	port, err := findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %v", err)
	}

	// Start Flask server
	logger.Infof("Starting Flask server on port %d...", port)
	server, err := startFlaskServer(tempDir, port, logger)
	if err != nil {
		return fmt.Errorf("failed to start Flask server: %v", err)
	}
	harness.RegisterTeardownFunc(func() { server.stop() })
	logger.Successf("Flask server started")

	// Create HTTP client
	client, err := newHTTPClient(server.baseURL)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 2. Test application startup - GET /
	logger.Infof("Testing application startup...")
	resp, _, err := client.get("/")
	if err != nil {
		return fmt.Errorf("failed to connect to application: %v", err)
	}
	// Should redirect to /login (302) or show login page
	if resp.StatusCode != 200 && resp.StatusCode != 302 {
		return fmt.Errorf("application startup failed, expected 200 or 302, got %d", resp.StatusCode)
	}
	logger.Successf("application starts")

	// 3. Test register page - GET /register
	logger.Infof("Testing register page...")
	resp, body, err := client.get("/register")
	if err != nil {
		return fmt.Errorf("failed to get register page: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("register page returned %d, expected 200", resp.StatusCode)
	}
	// Check for form fields
	if !containsFormField(body, "username") {
		return fmt.Errorf("register page missing username field")
	}
	if !containsFormField(body, "password") {
		return fmt.Errorf("register page missing password field")
	}
	if !containsFormField(body, "confirmation") {
		return fmt.Errorf("register page missing confirmation field")
	}
	logger.Successf("register page has required fields")

	// 4. Test registration with empty username
	logger.Infof("Testing registration with empty username...")
	resp, _, err = client.postForm("/register", url.Values{
		"username":     {""},
		"password":     {"password123"},
		"confirmation": {"password123"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to register: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("empty username should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("empty username rejected")

	// 5. Test registration with password mismatch
	logger.Infof("Testing registration with password mismatch...")
	resp, _, err = client.postForm("/register", url.Values{
		"username":     {"testuser"},
		"password":     {"password123"},
		"confirmation": {"differentpassword"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to register: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("password mismatch should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("password mismatch rejected")

	// 6. Test successful registration
	logger.Infof("Testing successful registration...")
	resp, _, err = client.postForm("/register", url.Values{
		"username":     {"testuser"},
		"password":     {"password123"},
		"confirmation": {"password123"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to register: %v", err)
	}
	// Should redirect to / (302 or 303)
	if resp.StatusCode != 302 && resp.StatusCode != 303 && resp.StatusCode != 200 {
		return fmt.Errorf("successful registration should redirect, got %d", resp.StatusCode)
	}
	logger.Successf("registration succeeds")

	// 7. Test duplicate username rejection
	logger.Infof("Testing duplicate username rejection...")
	// Create a new client to avoid session issues
	client2, _ := newHTTPClient(server.baseURL)
	resp, _, err = client2.postForm("/register", url.Values{
		"username":     {"testuser"},
		"password":     {"password456"},
		"confirmation": {"password456"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to register: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("duplicate username should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("duplicate username rejected")

	// 8. Test login page - GET /login
	logger.Infof("Testing login page...")
	client3, _ := newHTTPClient(server.baseURL)
	resp, body, err = client3.get("/login")
	if err != nil {
		return fmt.Errorf("failed to get login page: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("login page returned %d, expected 200", resp.StatusCode)
	}
	if !containsFormField(body, "username") {
		return fmt.Errorf("login page missing username field")
	}
	if !containsFormField(body, "password") {
		return fmt.Errorf("login page missing password field")
	}
	logger.Successf("login page has required fields")

	// 9. Test successful login
	logger.Infof("Testing successful login...")
	resp, _, err = client3.postForm("/login", url.Values{
		"username": {"testuser"},
		"password": {"password123"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to login: %v", err)
	}
	if resp.StatusCode != 302 && resp.StatusCode != 303 && resp.StatusCode != 200 {
		return fmt.Errorf("successful login should redirect, got %d", resp.StatusCode)
	}
	logger.Successf("login succeeds")

	// 10. Test quote page - GET /quote
	logger.Infof("Testing quote page...")
	resp, body, err = client3.get("/quote")
	if err != nil {
		return fmt.Errorf("failed to get quote page: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("quote page returned %d, expected 200", resp.StatusCode)
	}
	if !containsFormField(body, "symbol") {
		return fmt.Errorf("quote page missing symbol field")
	}
	logger.Successf("quote page has symbol field")

	// 11. Test quote with invalid symbol
	logger.Infof("Testing quote with invalid symbol...")
	resp, _, err = client3.postForm("/quote", url.Values{
		"symbol": {"ZZZZ"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to quote: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("invalid symbol should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("invalid symbol rejected")

	// 12. Test quote with blank symbol
	logger.Infof("Testing quote with blank symbol...")
	resp, _, err = client3.postForm("/quote", url.Values{
		"symbol": {""},
	})
	if err != nil {
		return fmt.Errorf("failed to post to quote: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("blank symbol should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("blank symbol rejected")

	// 13. Test quote with valid symbol
	logger.Infof("Testing quote with valid symbol...")
	resp, body, err = client3.postForm("/quote", url.Values{
		"symbol": {"AAAA"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to quote: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("valid symbol should return 200, got %d", resp.StatusCode)
	}
	// Should show price $28.00
	if !strings.Contains(body, "28.00") {
		return fmt.Errorf("quote response should contain price 28.00")
	}
	logger.Successf("valid quote returns price")

	// 14. Test buy page - GET /buy
	logger.Infof("Testing buy page...")
	resp, body, err = client3.get("/buy")
	if err != nil {
		return fmt.Errorf("failed to get buy page: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("buy page returned %d, expected 200", resp.StatusCode)
	}
	if !containsFormField(body, "symbol") {
		return fmt.Errorf("buy page missing symbol field")
	}
	if !containsFormField(body, "shares") {
		return fmt.Errorf("buy page missing shares field")
	}
	logger.Successf("buy page has required fields")

	// 15. Test buy with invalid symbol
	logger.Infof("Testing buy with invalid symbol...")
	resp, _, err = client3.postForm("/buy", url.Values{
		"symbol": {"ZZZZ"},
		"shares": {"4"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to buy: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("invalid symbol should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("buy with invalid symbol rejected")

	// 16. Test buy with invalid shares
	logger.Infof("Testing buy with invalid shares...")
	for _, invalidShares := range []string{"-1", "1.5", "foo"} {
		resp, _, err = client3.postForm("/buy", url.Values{
			"symbol": {"AAAA"},
			"shares": {invalidShares},
		})
		if err != nil {
			return fmt.Errorf("failed to post to buy: %v", err)
		}
		if resp.StatusCode != 400 {
			return fmt.Errorf("invalid shares '%s' should return 400, got %d", invalidShares, resp.StatusCode)
		}
	}
	logger.Successf("buy with invalid shares rejected")

	// 17. Test successful buy
	logger.Infof("Testing successful buy...")
	resp, _, err = client3.postForm("/buy", url.Values{
		"symbol": {"AAAA"},
		"shares": {"4"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to buy: %v", err)
	}
	if resp.StatusCode != 302 && resp.StatusCode != 303 && resp.StatusCode != 200 {
		return fmt.Errorf("successful buy should redirect, got %d", resp.StatusCode)
	}
	logger.Successf("buy succeeds")

	// 18. Verify portfolio after buy
	logger.Infof("Verifying portfolio after buy...")
	resp, body, err = client3.get("/")
	if err != nil {
		return fmt.Errorf("failed to get portfolio: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("portfolio page returned %d, expected 200", resp.StatusCode)
	}
	// Should show AAAA shares and value ($28 * 4 = $112)
	if !strings.Contains(body, "AAAA") {
		return fmt.Errorf("portfolio should show AAAA")
	}
	if !strings.Contains(body, "112") {
		return fmt.Errorf("portfolio should show value 112.00 (4 shares * $28)")
	}
	// Cash should be $10000 - $112 = $9888
	if !strings.Contains(body, "9,888") && !strings.Contains(body, "9888") {
		return fmt.Errorf("portfolio should show cash 9888.00")
	}
	logger.Successf("portfolio shows correct values after buy")

	// 19. Test sell page - GET /sell
	logger.Infof("Testing sell page...")
	resp, body, err = client3.get("/sell")
	if err != nil {
		return fmt.Errorf("failed to get sell page: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("sell page returned %d, expected 200", resp.StatusCode)
	}
	if !containsFormField(body, "symbol") && !containsSelectField(body, "symbol") {
		return fmt.Errorf("sell page missing symbol field")
	}
	if !containsFormField(body, "shares") {
		return fmt.Errorf("sell page missing shares field")
	}
	logger.Successf("sell page has required fields")

	// 20. Test sell with too many shares
	logger.Infof("Testing sell with too many shares...")
	resp, _, err = client3.postForm("/sell", url.Values{
		"symbol": {"AAAA"},
		"shares": {"8"}, // Only have 4
	})
	if err != nil {
		return fmt.Errorf("failed to post to sell: %v", err)
	}
	if resp.StatusCode != 400 {
		return fmt.Errorf("selling too many shares should return 400, got %d", resp.StatusCode)
	}
	logger.Successf("sell with too many shares rejected")

	// 21. Test successful sell
	logger.Infof("Testing successful sell...")
	resp, _, err = client3.postForm("/sell", url.Values{
		"symbol": {"AAAA"},
		"shares": {"2"},
	})
	if err != nil {
		return fmt.Errorf("failed to post to sell: %v", err)
	}
	if resp.StatusCode != 302 && resp.StatusCode != 303 && resp.StatusCode != 200 {
		return fmt.Errorf("successful sell should redirect, got %d", resp.StatusCode)
	}
	logger.Successf("sell succeeds")

	// 22. Verify portfolio after sell
	logger.Infof("Verifying portfolio after sell...")
	resp, body, err = client3.get("/")
	if err != nil {
		return fmt.Errorf("failed to get portfolio: %v", err)
	}
	// Should now have 2 shares worth $56
	if !strings.Contains(body, "56") {
		return fmt.Errorf("portfolio should show value 56.00 (2 shares * $28)")
	}
	// Cash should be $9888 + $56 = $9944
	if !strings.Contains(body, "9,944") && !strings.Contains(body, "9944") {
		return fmt.Errorf("portfolio should show cash 9944.00")
	}
	logger.Successf("portfolio shows correct values after sell")

	// 23. Test history page
	logger.Infof("Testing history page...")
	resp, body, err = client3.get("/history")
	if err != nil {
		return fmt.Errorf("failed to get history page: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("history page returned %d, expected 200", resp.StatusCode)
	}
	// Should show transactions
	if !strings.Contains(body, "AAAA") {
		return fmt.Errorf("history should show AAAA transactions")
	}
	logger.Successf("history page shows transactions")

	logger.Successf("All tests passed!")
	return nil
}

// containsFormField checks if HTML contains an input field with the given name
func containsFormField(html, fieldName string) bool {
	// Match input fields: <input ... name="fieldName" ...>
	pattern := fmt.Sprintf(`<input[^>]*name=["']%s["'][^>]*>`, fieldName)
	matched, _ := regexp.MatchString(pattern, html)
	return matched
}

// containsSelectField checks if HTML contains a select field with the given name
func containsSelectField(html, fieldName string) bool {
	// Match select fields: <select ... name="fieldName" ...>
	pattern := fmt.Sprintf(`<select[^>]*name=["']%s["'][^>]*>`, fieldName)
	matched, _ := regexp.MatchString(pattern, html)
	return matched
}

// findAvailablePort finds an available TCP port
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// copyDir copies a directory recursively, symlinking .venv for speed
func copyDir(src, dst string) error {
	// First, check if .venv exists and create a symlink for it
	venvSrc := filepath.Join(src, ".venv")
	if info, err := os.Stat(venvSrc); err == nil && info.IsDir() {
		venvDst := filepath.Join(dst, ".venv")
		if err := os.Symlink(venvSrc, venvDst); err != nil {
			return fmt.Errorf("failed to symlink .venv: %v", err)
		}
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// Skip flask_session directory
		if strings.Contains(path, "flask_session") {
			return nil
		}

		// Skip .venv directory since we already symlinked it
		if info.IsDir() && info.Name() == ".venv" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

// resetDatabase resets the finance.db to initial state
func resetDatabase(dbPath string) error {
	// Remove existing db
	os.Remove(dbPath)

	// Create new database with schema
	db, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	db.Close()

	// Use sqlite3 to create tables
	cmd := exec.Command("sqlite3", dbPath)
	cmd.Stdin = strings.NewReader(`
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    username TEXT NOT NULL,
    hash TEXT NOT NULL,
    cash NUMERIC NOT NULL DEFAULT 10000.00
);
CREATE UNIQUE INDEX username ON users (username);
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    shares INTEGER NOT NULL,
    price REAL NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`)
	return cmd.Run()
}
