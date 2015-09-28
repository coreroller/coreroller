package api

import "time"

// Team represents a CoreRoller team.
type Team struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedTs time.Time `db:"created_ts"`
}

// AddTeam registers a team.
func (api *API) AddTeam(team *Team) (*Team, error) {
	var err error

	if team.ID != "" {
		err = api.dbR.InsertInto("team").Whitelist("id", "name").Record(team).Returning("*").QueryStruct(team)
	} else {
		err = api.dbR.InsertInto("team").Whitelist("name").Record(team).Returning("*").QueryStruct(team)
	}

	return team, err
}
