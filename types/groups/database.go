package groups

import (
	"sort"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
)

var (
	db  database.Database
	log = utils.Log.WithField("type", "group")
)

func SetDB(database database.Database) {
	db = database.Model(&Group{})
}

func (g *Group) Validate() error {
	if g.Name == "" {
		return errors.New("group name is empty")
	}
	return nil
}

func (g *Group) AfterFind() {
	metrics.Query("group", "find")
}

func (g *Group) AfterUpdate() {
	metrics.Query("group", "update")
}

func (g *Group) AfterDelete() {
	metrics.Query("group", "delete")
}

func (g *Group) BeforeUpdate() error {
	return g.Validate()
}

func (g *Group) BeforeCreate() error {
	return g.Validate()
}

func (g *Group) AfterCreate() {
	metrics.Query("group", "create")
}

func Find(id int64) (*Group, error) {
	var group Group
	q := db.Where("id = ?", id).Find(&group)
	if q.Error() != nil {
		return nil, errors.Missing(group, id)
	}
	return &group, q.Error()
}

func FindByName(name string) (*Group, error) {
	var group Group
	q := db.Where("name = ?", name).Find(&group)
	if q.Error() != nil {
		return nil, errors.Missing(group, name)
	}
	return &group, q.Error()
}

func All() []*Group {
	var groups []*Group
	db.Find(&groups)
	return groups
}

func AllWith(limit, offset int64) []*Group {
	var groups []*Group
	db.Limit(int(limit)).Offset(int(offset)).Find(&groups)
	return groups
}

func (g *Group) Create() error {
	q := db.Create(g)
	return q.Error()
}

func (g *Group) Update() error {
	q := db.Update(g)
	return q.Error()
}

func (g *Group) Delete() error {
	q := db.Delete(g)
	return q.Error()
}

// SelectGroups returns all groups
func SelectGroups(includeAll bool, auth bool) []*Group {
	var validGroups []*Group

	all := All()
	if includeAll {
		sort.Sort(GroupOrder(all))
		return all
	}

	for _, g := range all {
		if !g.Public.Bool {
			if auth {
				validGroups = append(validGroups, g)
			}
		} else {
			validGroups = append(validGroups, g)
		}
	}
	sort.Sort(GroupOrder(validGroups))
	return validGroups
}

// SelectGroupsWith returns all groups
func SelectGroupsWith(limit, offset int64, includeAll bool, auth bool) []*Group {
	var validGroups []*Group

	all := AllWith(limit, offset)
	if includeAll {
		sort.Sort(GroupOrder(all))
		return all
	}

	for _, g := range all {
		if !g.Public.Bool {
			if auth {
				validGroups = append(validGroups, g)
			}
		} else {
			validGroups = append(validGroups, g)
		}
	}
	sort.Sort(GroupOrder(validGroups))
	return validGroups
}
