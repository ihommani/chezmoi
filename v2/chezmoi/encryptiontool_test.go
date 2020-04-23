package chezmoi

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

type testEncryptionTool struct {
	key byte
}

var _ EncryptionTool = &testEncryptionTool{}

type testEncryptionToolOption func(*testEncryptionTool)

func withKey(key byte) testEncryptionToolOption {
	return func(t *testEncryptionTool) {
		t.key = key
	}
}

func newTestEncryptionTool(options ...testEncryptionToolOption) *testEncryptionTool {
	t := &testEncryptionTool{
		key: byte(rand.Int() + 1),
	}
	for _, option := range options {
		option(t)
	}
	return t
}

func (t *testEncryptionTool) Decrypt(filenameHint string, ciphertext []byte) ([]byte, error) {
	return t.xorWithKey(ciphertext), nil
}

func (t *testEncryptionTool) DecryptToFile(filenameHint string, ciphertext []byte) (filename string, cleanupFunc CleanupFunc, err error) {
	tempDir, err := ioutil.TempDir("", "chezmoi-test-decrypt")
	if err != nil {
		return
	}
	cleanupFunc = func() error {
		return os.RemoveAll(tempDir)
	}

	filename = filepath.Join(tempDir, filepath.Base(filenameHint))
	if err = ioutil.WriteFile(filename, t.xorWithKey(ciphertext), 0o600); err != nil {
		err = multierr.Append(err, cleanupFunc())
		return
	}

	return
}

func (t *testEncryptionTool) Encrypt(plaintext []byte) ([]byte, error) {
	return t.xorWithKey(plaintext), nil
}

func (t *testEncryptionTool) EncryptFile(filename string) ([]byte, error) {
	plaintext, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return t.xorWithKey(plaintext), nil
}

func (t *testEncryptionTool) xorWithKey(input []byte) []byte {
	output := make([]byte, 0, len(input))
	for _, b := range input {
		output = append(output, b^t.key)
	}
	return output
}

func TestTestEncryptionToolDecryptToFile(t *testing.T) {
	et := newTestEncryptionTool()
	expectedPlaintext := []byte("secret")

	actualCiphertext, err := et.Encrypt(expectedPlaintext)
	require.NoError(t, err)
	assert.NotEqual(t, expectedPlaintext, actualCiphertext)

	filenameHint := "filename.txt"
	filename, cleanup, err := et.DecryptToFile(filenameHint, actualCiphertext)
	require.NoError(t, err)
	assert.True(t, strings.Contains(filename, filenameHint))
	assert.NotNil(t, cleanup)
	defer func() {
		assert.NoError(t, cleanup())
	}()

	actualPlaintext, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, expectedPlaintext, actualPlaintext)
}

func TestTestEncryptionToolEncryptDecrypt(t *testing.T) {
	et := newTestEncryptionTool()
	expectedPlaintext := []byte("secret")

	actualCiphertext, err := et.Encrypt(expectedPlaintext)
	require.NoError(t, err)
	assert.NotEqual(t, expectedPlaintext, actualCiphertext)

	actualPlaintext, err := et.Decrypt("", actualCiphertext)
	require.NoError(t, err)
	assert.Equal(t, expectedPlaintext, actualPlaintext)
}

func TestTestEncryptionToolEncryptFile(t *testing.T) {
	et := newTestEncryptionTool()

	tempFile, err := ioutil.TempFile("", "chezmoi-test-encyrption-tool")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tempFile.Name()))
	}()
	expectedPlaintext := []byte("secret")
	require.NoError(t, ioutil.WriteFile(tempFile.Name(), expectedPlaintext, 0o600))

	actualCiphertext, err := et.EncryptFile(tempFile.Name())
	require.NoError(t, err)
	assert.NotEqual(t, expectedPlaintext, actualCiphertext)

	actualPlaintext, err := et.Decrypt("", actualCiphertext)
	require.NoError(t, err)
	assert.Equal(t, expectedPlaintext, actualPlaintext)
}
