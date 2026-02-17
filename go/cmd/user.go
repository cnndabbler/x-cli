package cmd

import (
	"x-cli/internal/output"
)

type UserCmd struct {
	Get       UserGetCmd       `cmd:"" help:"Look up a user profile"`
	Timeline  UserTimelineCmd  `cmd:"" help:"Fetch a user's recent tweets"`
	Followers UserFollowersCmd `cmd:"" help:"List a user's followers"`
	Following UserFollowingCmd `cmd:"" help:"List who a user follows"`
}

// --- get ---

type UserGetCmd struct {
	Username string `arg:"" help:"Username (with or without @)"`
}

func (c *UserGetCmd) Run(g *Globals) error {
	client := newClient()
	uname := stripAt(c.Username)
	data, err := client.GetUser(uname)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "@"+uname, g.Verbose)
	return nil
}

// --- timeline ---

type UserTimelineCmd struct {
	Username   string `arg:"" help:"Username (with or without @)"`
	MaxResults int    `name:"max" default:"10" help:"Max results (5-100)"`
}

func (c *UserTimelineCmd) Run(g *Globals) error {
	client := newClient()
	uname := stripAt(c.Username)
	userData, err := client.GetUser(uname)
	if err != nil {
		return err
	}
	d, _ := userData["data"].(map[string]any)
	uid, _ := d["id"].(string)
	data, err := client.GetTimeline(uid, c.MaxResults, "")
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "@"+uname+" timeline", g.Verbose)
	return nil
}

// --- followers ---

type UserFollowersCmd struct {
	Username   string `arg:"" help:"Username (with or without @)"`
	MaxResults int    `name:"max" default:"100" help:"Max results (1-1000)"`
}

func (c *UserFollowersCmd) Run(g *Globals) error {
	client := newClient()
	uname := stripAt(c.Username)
	userData, err := client.GetUser(uname)
	if err != nil {
		return err
	}
	d, _ := userData["data"].(map[string]any)
	uid, _ := d["id"].(string)
	data, err := client.GetFollowers(uid, c.MaxResults)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "@"+uname+" followers", g.Verbose)
	return nil
}

// --- following ---

type UserFollowingCmd struct {
	Username   string `arg:"" help:"Username (with or without @)"`
	MaxResults int    `name:"max" default:"100" help:"Max results (1-1000)"`
}

func (c *UserFollowingCmd) Run(g *Globals) error {
	client := newClient()
	uname := stripAt(c.Username)
	userData, err := client.GetUser(uname)
	if err != nil {
		return err
	}
	d, _ := userData["data"].(map[string]any)
	uid, _ := d["id"].(string)
	data, err := client.GetFollowing(uid, c.MaxResults)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "@"+uname+" following", g.Verbose)
	return nil
}
