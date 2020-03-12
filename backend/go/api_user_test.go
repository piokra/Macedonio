package macedonio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createRequestAndRecorder(method string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	req, _ := http.NewRequest(method, httptest.DefaultRemoteAddr, body)
	recorder := httptest.NewRecorder()
	return recorder, req
}

func createUser(username, password, email string) error {
	newUser := User{
		Username:   username,
		Email:      email,
		Password:   password,
		UserStatus: 0,
	}

	data, err := json.Marshal(&newUser)

	if err != nil {
		return err
	}

	recorder, req := createRequestAndRecorder("POST", bytes.NewReader(data))

	CreateUser(recorder, req)

	if recorder.Code != 200 {
		return fmt.Errorf("failed to create user with code %d", recorder.Code)
	}

	return nil
}

func getUserByUsername(username, token64 string) (User, error) {
	recorder, req := createRequestAndRecorder("POST", nil)
	var user User

	query := req.URL.Query()
	query.Set(USERNAME, username)
	req.URL.RawQuery = query.Encode()

	if token64 != "" {
		cookie := http.Cookie{
			Name:    TOKEN,
			Value:   token64,
			Path:    "",
			Domain:  "",
			Expires: time.Now().Add(time.Hour),
		}
		req.AddCookie(&cookie)
	}

	GetUserByName(recorder, req)

	if recorder.Code != 200 {
		return user, fmt.Errorf("failed to get user by name. error code: %d", recorder.Code)
	}

	data, err := ioutil.ReadAll(recorder.Body)

	if err != nil {
		return user, fmt.Errorf("failed to read data %v", err)
	}

	err = json.Unmarshal(data, &user)

	if err != nil {
		return user, fmt.Errorf("failed to unmarshall data err: %v", err)
	}

	return user, nil
}

func deleteUser(username, token64 string) error {
	recorder, req := createRequestAndRecorder("POST", nil)
	cookie := http.Cookie{
		Name:    TOKEN,
		Value:   token64,
		Expires: time.Now().Add(time.Hour),
		Secure:  true,
	}

	query := req.URL.Query()
	query.Set(USERNAME, username)
	req.URL.RawQuery = query.Encode()

	req.AddCookie(&cookie)
	DeleteUser(recorder, req)
	if recorder.Code != 200 {
		return fmt.Errorf("failed to delete user. code: %d", recorder.Code)
	}
	return nil
}

func logoutUser(token64 string) error {
	recorder, req := createRequestAndRecorder("POST", nil)
	cookie := http.Cookie{
		Name:    TOKEN,
		Value:   token64,
		Expires: time.Now().Add(time.Hour),
		Secure:  true,
	}
	req.AddCookie(&cookie)
	LogoutUser(recorder, req)

	if recorder.Code != 200 {
		return fmt.Errorf("failed to logout user with code %d", recorder.Code)
	}
	return nil
}

func loginUser(username, password string) (string, error) {
	var token string

	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	data, _ := json.Marshal(&loginData)

	recorder, req := createRequestAndRecorder("POST", bytes.NewReader(data))

	LoginUser(recorder, req)

	if recorder.Code != 200 {
		return token, fmt.Errorf("Status was not ok %d", recorder.Code)
	}

	if recorder.Result() != nil {
		cookies := ExtractCookies(recorder.Result().Cookies())
		token, ok := cookies[TOKEN]

		if !ok {
			return "", fmt.Errorf("no token found in cookies")
		}

		return token.Value, nil
	}

	return "", fmt.Errorf("no token found")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n uint) string {
	ans := make([]rune, n)
	for i := range ans {
		ans[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(ans)
}

func randomUsernamePasswordEmail() (string, string, string) {
	return randomString(16), randomString(16), randomString(16) + "@" + randomString(2) + "." + randomString(2)
}

func TestUserAPI(t *testing.T) {
	err := InitDBHandle()
	if err != nil {
		panic(err)
	}
	GetDBHandle().Exec("TRUNCATE TABLE db_users")
	AutoMigrateUserSchemas()

	t.Run("User creation", func(t *testing.T) {
		err := createUser("pan", "piotr", "pp@p.p")
		if err != nil {
			t.Errorf("failed to create user with error: %v", err)
		}

		for i := 0; i < 30; i++ {
			err := createUser(randomString(16),
				randomString(16), randomString(16)+"@"+randomString(2)+"."+randomString(2))
			if err != nil {
				t.Errorf("failed to create user with error: %v", err)
			}
		}
	})

	t.Run("Login with created user", func(t *testing.T) {
		username, password, email := randomUsernamePasswordEmail()

		err := createUser(username, password, email)
		if err != nil {
			t.Fatal(err)
		}

		token, err := loginUser(username, password)

		if err != nil {
			t.Fatal(err)
		}

		user, err := TokenToUser(token)

		if err != nil {
			t.Fatal(err)
		}

		if user.Username != username {
			t.Errorf("Wrong username. Expected: %s. Got: %s. All: %+v", username, user.Username, user)
		}

		if bytes.Contains(user.SaltedPasswordHash, []byte(password)) {
			t.Errorf("Unsalted/hashed password")
		}
	})

	t.Run("Login with unknown user", func(t *testing.T) {
		token, err := loginUser("unknown", "unknown")
		if err == nil {
			user, err := TokenToUser(token)
			if err != nil {
				t.Fatalf("Logged in with unknown user but failed to convert the token to user")
			} else {
				t.Errorf("Loggined with invalid credentials (%s:%s). Logged user: %s",
					"unknown", "unknown", user.Username)
			}

		}
	})

	t.Run("Logout with valid token", func(t *testing.T) {
		username, password, email := randomString(17), randomString(17), randomString(17)+"@xx.xx"
		err := createUser(username, password, email)

		if err != nil {
			t.Error("failed to create user")
		}

		token, err := loginUser(username, password)
		if err != nil {
			t.Fatal(err)
		}

		err = logoutUser(token)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Create and delete users", func(t *testing.T) {
		type Args struct {
			username, password, email string
		}

		var args [30]Args
		for i := range args {
			args[i].username = randomString(16)
			args[i].password = randomString(16)
			args[i].email = randomString(16) + "@" + randomString(2) + "." + randomString(2)
			err := createUser(args[i].username, args[i].password, args[i].email)
			if err != nil {
				t.Error(err)
			}

		}

		permutations := rand.Perm(30)
		for _, p := range permutations {
			token, err := loginUser(args[p].username, args[p].password)

			if err != nil {
				t.Errorf("failed to login with %+v", args[p])
				continue
			}

			err = deleteUser(args[p].username, token)

			if err != nil {
				t.Errorf("failed to delete %+v", args[p])
				continue
			}
		}

	})

	t.Run("Get logged in user", func(t *testing.T) {
		username, password, email := randomUsernamePasswordEmail()
		err := createUser(username, password, email)

		if err != nil {
			t.Fatal(err)
		}

		token, err := loginUser(username, password)

		if err != nil {
			t.Fatal(err)
		}

		user, err := getUserByUsername(username, token)

		if err != nil {
			t.Fatal(err)
		}

		if user.Username != username {
			t.Errorf("wrong username. expected: %s, got: %s", username, user.Username)
		}

		if user.Email != email {
			t.Errorf("wrong email. expected: %s, got: %s", email, user.Email)
		}

	})

	t.Run("Get user by username while not logged in", func(t *testing.T) {
		username, password, email := randomUsernamePasswordEmail()
		err := createUser(username, password, email)

		if err != nil {
			t.Fatal(err)
		}

		token, err := loginUser(username, password)

		if err != nil {
			t.Fatal(err)
		}

		err = logoutUser(token)
		if err != nil {
			t.Errorf("failed to log out. reason: %v", err)
		}

		user, err := getUserByUsername(username, token)

		if err != nil {
			t.Fatal(err)
		}

		if user.Username != username {
			t.Errorf("wrong username. expected: %s, got: %s", username, user.Username)
		}

		if user.Email != "" {
			t.Errorf("wrong email. expected: <empty>, got: %s", user.Email)
		}

	})

	t.Run("Login, Logout, Delete user", func(t *testing.T) {
		username, password, email := randomUsernamePasswordEmail()
		err := createUser(username, password, email)
		if err != nil {
			t.Fatal(err)
		}

		token64, err := loginUser(username, password)

		if err != nil {
			t.Fatal(err)
		}

		err = logoutUser(token64)

		if err != nil {
			t.Fatal(err)
		}

		err = deleteUser(username, token64)

		if err == nil {
			t.Fatal("Deleted user after logging out")
		}
	})
}
