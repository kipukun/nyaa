package cookies

import "github.com/gin-gonic/gin"

const (
	// CookieName : Name of cookie
	CookieName = "session"

	// UserContextKey : key for user context
	UserContextKey = "nyaapantsu.user"
)
// CreateUserAuthentication creates user authentication.
func CreateUserAuthentication(c *gin.Context, form *formStruct.LoginForm) (int, error) {
	user, status, err := api.CreateUserAuthentication(form)
	if err != nil {
		return status, err
	}
	status, err = SetCookieHandler(c, user)
	return status, err
}

// If you want to keep login cookies between restarts you need to make these permanent
var cookieHandler = securecookie.New(
	getOrGenerateKey(config.Conf.Cookies.HashKey, 64),
	getOrGenerateKey(config.Conf.Cookies.EncryptionKey, 32))

func getOrGenerateKey(key string, requiredLen int) []byte {
	data := []byte(key)
	if len(data) == 0 {
		data = securecookie.GenerateRandomKey(requiredLen)
	} else if len(data) != requiredLen {
		panic(fmt.Sprintf("failed to load cookie key. required key length is %d bytes and the provided key length is %d bytes.", requiredLen, len(data)))
	}
	return data
}

// Decode : Encoding & Decoding of the cookie value
func Decode(cookieValue string) (uint, error) {
	value := make(map[string]string)
	err := cookieHandler.Decode(CookieName, cookieValue, &value)
	if err != nil {
		return 0, err
	}
	timeInt, _ := strconv.ParseInt(value["t"], 10, 0)
	if timeHelper.IsExpired(time.Unix(timeInt, 0)) {
		return 0, errors.New("Cookie is expired")
	}
	ret, err := strconv.ParseUint(value["u"], 10, 0)
	return uint(ret), err
}

// Encode : Encoding of the cookie value
func Encode(userID uint, validUntil time.Time) (string, error) {
	value := map[string]string{
		"u": strconv.FormatUint(uint64(userID), 10),
		"t": strconv.FormatInt(validUntil.Unix(), 10),
	}
	return cookieHandler.Encode(CookieName, value)
}

// Clear : Erase cookie session
func Clear(c *gin.Context) {
	c.SetCookie(CookieName, "", -1, "/", getDomainName(), false, true)
}

// SetLogin sets the authentication cookie
func SetLogin(c *gin.Context, user models.User) (int, error) {
	maxAge := getMaxAge()
	validUntil := timeHelper.FewDurationLater(time.Duration(maxAge) * time.Second)
	encoded, err := Encode(user.ID, validUntil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	c.SetCookie(CookieName, encoded, maxAge, "/", getDomainName(), false, true)
	// also set response header for convenience
	c.Header("X-Auth-Token", encoded)
	return http.StatusOK, nil
}