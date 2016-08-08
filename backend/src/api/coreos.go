package api

import (
	"errors"

	"github.com/mgutz/logxi/v1"
)

const coreosAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"

var (
	coreosLogger = log.New("coreos")

	coreosGroups = map[string]string{
		"alpha":  "5b810680-e36a-4879-b98a-4f989e80b899",
		"beta":   "3fe10490-dd73-4b49-b72a-28ac19acfcdc",
		"stable": "9a2deb70-37be-4026-853f-bfdd6b347bbe",
	}

	coreosGroupsPopulated = false
)

// CoreosAppID getter for coreosAppID 
func CoreosAppID() string {
	return coreosAppID
}

// CoreosGroupID retrieves groupID from coreosGroups
func CoreosGroupID(group string) (string, bool) {
	uuid, ok := coreosGroups[group]

	return uuid, ok
}

func coreosGroupAdd(group string, uuid string) error {
	f := "GroupAdd"

	coreosLogger.Debug(f, "group", group, "uuid", uuid)

	if _, ok := coreosGroups[group]; ok {
		err := "Cannot add new CoreOS group as it already exists"
		coreosLogger.Warn(f, err)
		return errors.New(err)
	}

	coreosGroups[group] = uuid

	return nil
}

func coreosGroupDel(groupID string) error {
	f := "GroupDel"
	err := "Cannot delete non-existent CoreOS group"

	coreosLogger.Debug(f, "groupID", groupID)

	//Delete by group ID
	for group, uuid := range coreosGroups {
		if uuid == groupID {
			coreosLogger.Debug(f, "group", group, "groupID", groupID)
			delete(coreosGroups, group)
			return nil
		}
	}

	return errors.New(err)
}

func coreosGroupPopulate(api *API) {
	if coreosGroupsPopulated {
		return
	}

	f := "init"

	coreosLogger.Debug(f)

	groups, err := api.GetGroups(coreosAppID, 1, 1000)
	if err != nil {
		coreosLogger.Error(f, "api.GetGroups", err)
	}

	for _, group := range groups {
		if _, exist := coreosGroups[group.Name]; ! exist {
			coreosLogger.Debug(f, "Adding group", group, "uuid", group.ID)
			coreosGroups[group.Name] = group.ID
		}
	}

	coreosGroupsPopulated = true
}
