package kattest

import (
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func AssertStatusUnauthorized(t *testing.T, err error) {
	var ue *kathttpc.UnexpectedStatusCodeError
	assert.ErrorAs(t, err, &ue)
	assert.Equal(t, ue.StatusCode, http.StatusUnauthorized)
}

func AssertStatusNotFound(t *testing.T, err error) {
	var ue *kathttpc.UnexpectedStatusCodeError
	assert.ErrorAs(t, err, &ue)
	assert.Equal(t, ue.StatusCode, http.StatusNotFound)
}

func AssertStatusBadRequest(t *testing.T, err error) {
	var ue *kathttpc.UnexpectedStatusCodeError
	assert.ErrorAs(t, err, &ue)
	assert.Equal(t, ue.StatusCode, http.StatusBadRequest)
}

func AssertStatusConflict(t *testing.T, err error) {
	var ue *kathttpc.UnexpectedStatusCodeError
	assert.ErrorAs(t, err, &ue)
	assert.Equal(t, ue.StatusCode, http.StatusConflict)
}

func AssertStatusInternalServerError(t *testing.T, err error) {
	var ue *kathttpc.UnexpectedStatusCodeError
	assert.ErrorAs(t, err, &ue)
	assert.Equal(t, ue.StatusCode, http.StatusInternalServerError)
}
