package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/common/d2r"
	"d2rhl/internal/multiboxing/account"
)

func setupAccountDefaultRegions(accounts []account.Account, accountsFile string) {
	if len(accounts) == 0 {
		ui.infof("%s", lang.RegionDefaults.NoAccounts)
		ui.blankLine()
		return
	}

	_ = runMenu(func() {
		ui.headf("%s", lang.RegionDefaults.Title)
		ui.infof("%s", lang.RegionDefaults.Intro1)
		ui.infof("%s", lang.RegionDefaults.Intro2)
		ui.infof("%s", lang.RegionDefaults.Intro3)
		ui.blankLine()
		printAccountList(accounts, defaultRegionStatusLabel)
		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.RegionDefaults.OptAssign, "")
			options.option("2", lang.RegionDefaults.OptClear, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(choice string) error {
		switch choice {
		case "1":
			return assignDefaultRegions(accounts, accountsFile)
		case "2":
			return clearDefaultRegions(accounts, accountsFile)
		default:
			showInvalidInputAndPause()
			return nil
		}
	})
}

func assignDefaultRegions(accounts []account.Account, accountsFile string) error {
	return runMenu(func() {
		ui.headf("%s", lang.RegionDefaults.AssignModeTitle)
		ui.infof("%s", lang.RegionDefaults.AssignModeQuestion)
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.RegionDefaults.OptRegionToAccounts, "")
			options.option("2", lang.RegionDefaults.OptAccountToRegion, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(choice string) error {
		switch choice {
		case "1":
			return assignDefaultRegionsByRegion(accounts, accountsFile)
		case "2":
			return assignDefaultRegionsByAccount(accounts, accountsFile)
		default:
			showInvalidInputAndPause()
			return nil
		}
	})
}

func assignDefaultRegionsByRegion(accounts []account.Account, accountsFile string) error {
	regions := d2r.Regions
	var completed bool

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.RegionDefaults.AssignByRegionTitle)
			ui.menuBlock(func() {
				renderDefaultRegionOptions(regions)
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.RegionDefaults.AssignByRegionSelectPrompt)
		},
		func(input string) error {
			selected, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || selected < 1 || selected > len(regions) {
				showInvalidInputAndPause()
				return nil
			}

			region := regions[selected-1]
			var done bool
			innerErr := runMenuRead(
				func() {
					ui.headf("%s", lang.RegionDefaults.AssignByRegionAccountTitle)
					options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
						for i, acc := range accounts {
							menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.RegionDefaults.AccountComment, defaultRegionStatusLabel(acc)))
						}
					})
					ui.menuBlock(func() {
						options.render()
					})
				},
				func() (string, bool) {
					return ui.readInputf(lang.RegionDefaults.AssignByRegionAccountPrompt, region.Name)
				},
				func(input string) error {
					accountIndexes, err := parseSelectionInput(input, len(accounts))
					if err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
						return nil
					}

					ui.blankLine()
					ui.infof(lang.RegionDefaults.AssignByRegionAbout, region.Name)
					for _, idx := range accountIndexes {
						acc := accounts[idx]
						ui.rawlnf(lang.RegionDefaults.AccountItemFmt, idx+1, acc.DisplayName, acc.Email, defaultRegionStatusLabel(acc))
					}
					if !confirmChanges() {
						ui.infof("%s", lang.Common.Cancelled)
						ui.blankLine()
						return nil
					}

					if err := applyDefaultRegionAssignments(accounts, accountsFile, accountIndexes, region.Name); err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
						return nil
					}

					ui.successf(lang.RegionDefaults.AssignDone, region.Name)
					ui.blankLine()
					done = true
					return errNavDone
				},
			)
			if errors.Is(innerErr, ErrNavHome) {
				return ErrNavHome
			}
			if done {
				completed = true
				return errNavDone
			}
			return nil
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	if completed {
		return errNavDone
	}
	return nil
}

func assignDefaultRegionsByAccount(accounts []account.Account, accountsFile string) error {
	regions := d2r.Regions
	var completed bool

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.RegionDefaults.AssignByAccountTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, acc := range accounts {
					menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.RegionDefaults.AccountComment, defaultRegionStatusLabel(acc)))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.RegionDefaults.AssignByAccountSelectPrompt)
		},
		func(input string) error {
			selected, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || selected < 1 || selected > len(accounts) {
				showInvalidInputAndPause()
				return nil
			}

			accountIndex := selected - 1
			acc := accounts[accountIndex]
			var done bool
			innerErr := runMenuRead(
				func() {
					ui.headf("%s", lang.RegionDefaults.AssignByAccountRegionTitle)
					ui.infof(lang.RegionDefaults.AssignByAccountRegionPrompt, acc.DisplayName)
					ui.menuBlock(func() {
						renderDefaultRegionOptions(regions)
					})
				},
				func() (string, bool) {
					return ui.readInput()
				},
				func(input string) error {
					regionSelection, err := strconv.Atoi(strings.TrimSpace(input))
					if err != nil || regionSelection < 1 || regionSelection > len(regions) {
						showInvalidInputAndPause()
						return nil
					}

					region := regions[regionSelection-1]
					ui.blankLine()
					ui.infof(lang.RegionDefaults.AssignByAccountAbout, acc.DisplayName, region.Name)
					ui.rawlnf(lang.RegionDefaults.AccountItemFmt, selected, acc.DisplayName, acc.Email, defaultRegionStatusLabel(acc))
					if !confirmChanges() {
						ui.infof("%s", lang.Common.Cancelled)
						ui.blankLine()
						return nil
					}

					if err := applyDefaultRegionAssignments(accounts, accountsFile, []int{accountIndex}, region.Name); err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
						return nil
					}

					ui.successf(lang.RegionDefaults.AssignDone, region.Name)
					ui.blankLine()
					done = true
					return errNavDone
				},
			)
			if errors.Is(innerErr, ErrNavHome) {
				return ErrNavHome
			}
			if done {
				completed = true
				return errNavDone
			}
			return nil
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	if completed {
		return errNavDone
	}
	return nil
}

func clearDefaultRegions(accounts []account.Account, accountsFile string) error {
	assignedIndexes := assignedDefaultRegionAccountIndexes(accounts)
	if len(assignedIndexes) == 0 {
		showInfoAndPause(lang.RegionDefaults.ClearNoAssignments)
		return nil
	}

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.RegionDefaults.ClearTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, accountIndex := range assignedIndexes {
					acc := accounts[accountIndex]
					menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.RegionDefaults.AccountComment, defaultRegionStatusLabel(acc)))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.RegionDefaults.ClearPrompt)
		},
		func(input string) error {
			selectionIndexes, err := parseSelectionInput(input, len(assignedIndexes))
			if err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
				return nil
			}

			actualIndexes := make([]int, 0, len(selectionIndexes))
			ui.blankLine()
			ui.infof("%s", lang.RegionDefaults.ClearAbout)
			for _, idx := range selectionIndexes {
				accountIndex := assignedIndexes[idx]
				actualIndexes = append(actualIndexes, accountIndex)
				acc := accounts[accountIndex]
				ui.rawlnf(lang.RegionDefaults.AccountItemFmt, accountIndex+1, acc.DisplayName, acc.Email, defaultRegionStatusLabel(acc))
			}

			if !confirmChanges() {
				ui.infof("%s", lang.Common.Cancelled)
				ui.blankLine()
				return nil
			}

			if err := clearDefaultRegionAssignments(accounts, accountsFile, actualIndexes); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
				return nil
			}

			ui.successf("%s", lang.RegionDefaults.ClearDone)
			ui.blankLine()
			return errNavDone
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	return nil
}

func applyDefaultRegionAssignments(accounts []account.Account, accountsFile string, accountIndexes []int, regionName string) error {
	previous := make(map[int]string, len(accountIndexes))
	normalizedRegion := d2r.NormalizeRegionName(regionName)
	for _, idx := range accountIndexes {
		previous[idx] = accounts[idx].DefaultRegion
		accounts[idx].DefaultRegion = normalizedRegion
	}

	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, previousRegion := range previous {
			accounts[idx].DefaultRegion = previousRegion
		}
		return err
	}
	return nil
}

func clearDefaultRegionAssignments(accounts []account.Account, accountsFile string, accountIndexes []int) error {
	return applyDefaultRegionAssignments(accounts, accountsFile, accountIndexes, "")
}

func defaultRegionStatusLabel(acc account.Account) string {
	if strings.TrimSpace(acc.DefaultRegion) == "" {
		return lang.RegionDefaults.StatusUnassigned
	}
	return acc.DefaultRegion
}

func renderDefaultRegionOptions(regions []d2r.Region) {
	options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
		for i, region := range regions {
			menuOptions.option(strconv.Itoa(i+1), region.Name, region.Address)
		}
	})
	options.render()
}

func assignedDefaultRegionAccountIndexes(accounts []account.Account) []int {
	indexes := make([]int, 0, len(accounts))
	for i, acc := range accounts {
		if strings.TrimSpace(acc.DefaultRegion) == "" {
			continue
		}
		indexes = append(indexes, i)
	}
	return indexes
}
