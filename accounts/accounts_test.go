package accounts

import (
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/magiconair/properties/assert"
	"razor/core/types"
	"reflect"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	var path string
	var password string

	AccountUtilsInterface := AccountUtilsMock{}
	type args struct {
		account    accounts.Account
		accountErr error
	}
	tests := []struct {
		name          string
		args          args
		want          accounts.Account
		expectedFatal bool
	}{
		{
			name: "Test 1: When NewAccounts executes successfully",
			args: args{
				account: accounts.Account{Address: common.HexToAddress("0x000000000000000000000000000000000000dea1"),
					URL: accounts.URL{Scheme: "TestKeyScheme", Path: "test/key/path"},
				},
			},
			want: accounts.Account{Address: common.HexToAddress("0x000000000000000000000000000000000000dea1"),
				URL: accounts.URL{Scheme: "TestKeyScheme", Path: "test/key/path"},
			},
			expectedFatal: false,
		},
		{
			name: "Test 2: When there is an error in getting new account",
			args: args{
				accountErr: errors.New("account error"),
			},
			want:          accounts.Account{Address: common.HexToAddress("0x00")},
			expectedFatal: true,
		},
	}

	defer func() { log.ExitFunc = nil }()
	var fatal bool
	log.ExitFunc = func(int) { fatal = true }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewAccountMock = func(string, string) (accounts.Account, error) {
				return tt.args.account, tt.args.accountErr
			}

			fatal = false
			got := CreateAccount(path, password, AccountUtilsInterface)
			if tt.expectedFatal {
				assert.Equal(t, tt.expectedFatal, fatal)
			}
			if got.Address != tt.want.Address {
				t.Errorf("New address created, got = %v, want %v", got, tt.want.Address)
			}
		})
	}
}

func Test_getPrivateKeyFromKeystore(t *testing.T) {
	var password string
	var keystorePath string
	var privateKey *ecdsa.PrivateKey
	var jsonBytes []byte

	AccountUtilsInterface := AccountUtilsMock{}

	type args struct {
		jsonBytes    []byte
		jsonBytesErr error
		key          *keystore.Key
		keyErr       error
	}
	tests := []struct {
		name          string
		args          args
		want          *ecdsa.PrivateKey
		expectedFatal bool
	}{
		{
			name: "Test 1: When GetPrivateKey function executes successfully",
			args: args{
				jsonBytes: jsonBytes,
				key: &keystore.Key{
					PrivateKey: privateKey,
				},
			},
			want:          privateKey,
			expectedFatal: false,
		},
		{
			name: "Test 2: When there is an error in reading data from file",
			args: args{
				jsonBytesErr: errors.New("error in reading data"),
				key: &keystore.Key{
					PrivateKey: nil,
				},
			},
			want:          nil,
			expectedFatal: true,
		},
		{
			name: "Test 3: When there is an error in fetching private key",
			args: args{
				jsonBytes: jsonBytes,
				key: &keystore.Key{
					PrivateKey: nil,
				},
				keyErr: errors.New("private key error"),
			},
			want:          privateKey,
			expectedFatal: true,
		},
	}

	defer func() { log.ExitFunc = nil }()
	var fatal bool
	log.ExitFunc = func(int) { fatal = true }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReadFileMock = func(string) ([]byte, error) {
				return tt.args.jsonBytes, tt.args.jsonBytesErr
			}

			DecryptKeyMock = func([]byte, string) (*keystore.Key, error) {
				return tt.args.key, tt.args.keyErr
			}

			fatal = false
			got := getPrivateKeyFromKeystore(keystorePath, password, AccountUtilsInterface)
			if tt.expectedFatal {
				assert.Equal(t, tt.expectedFatal, fatal)
			}
			if got != tt.want {
				t.Errorf("Private key from GetPrivateKey, got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPrivateKey(t *testing.T) {
	var password string
	var keystorePath string
	var privateKey *ecdsa.PrivateKey

	AccountUtilsInterface := AccountUtilsMock{}

	accountsList := []accounts.Account{
		{Address: common.HexToAddress("0x000000000000000000000000000000000000dea1"),
			URL: accounts.URL{Scheme: "TestKeyScheme", Path: "test/key/path"},
		},
		{Address: common.HexToAddress("0x000000000000000000000000000000000000dea2"),
			URL: accounts.URL{Scheme: "TestKeyScheme", Path: "test/key/path"},
		},
	}

	type args struct {
		address    string
		accounts   []accounts.Account
		privateKey *ecdsa.PrivateKey
	}
	tests := []struct {
		name string
		args args
		want *ecdsa.PrivateKey
	}{
		{
			name: "Test 1: When input address is present in accountsList",
			args: args{
				address:    "0x000000000000000000000000000000000000dea1",
				accounts:   accountsList,
				privateKey: privateKey,
			},
			want: privateKey,
		},
		{
			name: "Test 2: When input address is npt present in accountsList",
			args: args{
				address:    "0x000000000000000000000000000000000000dea3",
				accounts:   accountsList,
				privateKey: privateKey,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AccountsMock = func(string) []accounts.Account {
				return tt.args.accounts
			}

			getPrivateKeyFromKeystoreMock = func(string, string, AccountInterface) *ecdsa.PrivateKey {
				return tt.args.privateKey
			}

			got := GetPrivateKey(tt.args.address, password, keystorePath, AccountUtilsInterface)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSign(t *testing.T) {
	var hash []byte
	var account types.Account
	var defaultPath string
	var privateKey *ecdsa.PrivateKey
	var signature []byte

	AccountUtilsInterface := AccountUtilsMock{}

	type args struct {
		privateKey   *ecdsa.PrivateKey
		signature    []byte
		signatureErr error
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "Test 1: When Sign function returns no error",
			args: args{
				privateKey:   privateKey,
				signature:    signature,
				signatureErr: nil,
			},
			want:    signature,
			wantErr: nil,
		},
		{
			name: "Test 2: When Sign function returns error",
			args: args{
				privateKey:   privateKey,
				signatureErr: nil,
			},
			want:    nil,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetPrivateKeyMock = func(string, string, string, AccountInterface) *ecdsa.PrivateKey {
				return tt.args.privateKey
			}

			SignMock = func([]byte, *ecdsa.PrivateKey) ([]byte, error) {
				return tt.args.signature, tt.args.signatureErr
			}

			got, err := Sign(hash, account, defaultPath, AccountUtilsInterface)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sign() got = %v, want %v", got, tt.want)
			}

			if err == nil || tt.wantErr == nil {
				if err != tt.wantErr {
					t.Errorf("Error for Sign function, got = %v, want = %v", err, tt.wantErr)
				}
			} else {
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("Error for Sign function, got = %v, want = %v", err, tt.wantErr)
				}
			}
		})
	}
}
