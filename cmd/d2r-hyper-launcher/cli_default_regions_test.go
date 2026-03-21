package main

import (
	"path/filepath"
	"testing"

	"d2rhl/internal/multiboxing/account"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRegionStatusLabel(t *testing.T) {
	assert.Equal(t, lang.RegionDefaults.StatusUnassigned, defaultRegionStatusLabel(account.Account{}))
	assert.Equal(t, "EU", defaultRegionStatusLabel(account.Account{DefaultRegion: "EU"}))
}

func TestAssignDefaultRegionsByAccountPersistsSelection(t *testing.T) {
	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err := account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	withTestInput(t, "1\n2\n\n", func() {
		err = assignDefaultRegionsByAccount(accounts, accountsFile)
	})

	assert.ErrorIs(t, err, errNavDone)
	assert.Equal(t, "EU", accounts[0].DefaultRegion)

	reloaded, loadErr := account.LoadAccounts(accountsFile)
	assert.NoError(t, loadErr)
	assert.Equal(t, "EU", reloaded[0].DefaultRegion)
}

func TestClearDefaultRegionsPersistsSelection(t *testing.T) {
	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", DefaultRegion: "NA"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err := account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	withTestInput(t, "1\n\n", func() {
		err = clearDefaultRegions(accounts, accountsFile)
	})

	assert.NoError(t, err)
	assert.Equal(t, "", accounts[0].DefaultRegion)

	reloaded, loadErr := account.LoadAccounts(accountsFile)
	assert.NoError(t, loadErr)
	assert.Equal(t, "", reloaded[0].DefaultRegion)
}
