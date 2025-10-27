package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	uid := uuid.New()
	tokenSecret := "asdfasdfasdf"
	expiresIn := time.Second * 15
	jwt, err := MakeJWT(uid, []byte(tokenSecret), expiresIn)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(jwt)
	vuid, err := ValidateJWT(jwt, []byte(tokenSecret))
	if err != nil {
		t.Error(err)
	}
	if vuid != uid {
		t.Errorf("vuid != uid")
	}
}
