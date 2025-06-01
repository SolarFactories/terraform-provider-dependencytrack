// Bootstrap authentication and initial project for use within Pipeline.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	dtrack "github.com/DependencyTrack/client-go"
)

var (
	Host            = os.Getenv("HOST")
	Username        = os.Getenv("USERNAME")
	CurrentPassword = os.Getenv("CURRENT_PASSWORD")
	Password        = os.Getenv("PASSWORD")
	TeamName        = os.Getenv("TEAM_NAME")
	ProjectName     = os.Getenv("PROJECT_NAME")
	ProjectVersion  = os.Getenv("PROJECT_VERSION")
)

func getToken(ctx context.Context) (string, error) {
	client, err := dtrack.NewClient(Host)
	if err != nil {
		return "", errors.New("Unable to create client, from: " + err.Error())
	}

	err = client.User.ForceChangePassword(ctx, Username, CurrentPassword, Password)
	if err != nil {
		return "", errors.New("Unable to change password, from: " + err.Error())
	}

	token, err := client.User.Login(ctx, Username, Password)
	if err != nil {
		return "", errors.New("Unable to login as user, from: " + err.Error())
	}

	return token, nil
}

func createTeam(ctx context.Context, client *dtrack.Client) (dtrack.Team, error) {
	team, err := client.Team.Create(ctx, dtrack.Team{
		Name: TeamName,
	})
	if err != nil {
		return dtrack.Team{}, errors.New("Unable to create Team, from: " + err.Error())
	}

	err = dtrack.ForEach(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Permission], error) {
			return client.Permission.GetAll(ctx, po)
		},
		func(perm dtrack.Permission) error {
			_, permErr := client.Permission.AddPermissionToTeam(ctx, perm, team.UUID)
			if permErr != nil {
				return errors.New("Unable to grant permission to team, from: " + permErr.Error())
			}

			return nil
		},
	)
	if err != nil {
		return dtrack.Team{}, errors.New("Unable to fetch / grant permissions to team, from: " + err.Error())
	}

	return team, nil
}

func createProject(ctx context.Context, client *dtrack.Client) error {
	project, err := client.Project.Create(ctx, dtrack.Project{
		Name:    ProjectName,
		Version: ProjectVersion,
		Active:  true,
	})
	if err != nil {
		return errors.New("Unable to create project, from: " + err.Error())
	}

	_, err = client.ProjectProperty.Create(ctx, project.UUID, dtrack.ProjectProperty{
		Group:       "Group1",
		Name:        "Name1",
		Value:       "Value1",
		Type:        "STRING",
		Description: "Description1",
	})
	if err != nil {
		return errors.New("Unable to create project property 1, from: " + err.Error())
	}

	_, err = client.ProjectProperty.Create(ctx, project.UUID, dtrack.ProjectProperty{
		Group:       "Group2",
		Name:        "Name2",
		Value:       "2",
		Type:        "INTEGER",
		Description: "Description2",
	})
	if err != nil {
		return errors.New("Unable to create project property 2, from: " + err.Error())
	}

	return nil
}

func main() {
	ctx := context.Background()

	token, err := getToken(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	client, err := dtrack.NewClient(Host, dtrack.WithBearerToken(token))
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = client.Config.Update(ctx, dtrack.ConfigProperty{
		GroupName:   "vuln-source",
		Name:        "nvd.enabled",
		Value:       "false",
		Type:        "",
		Description: "",
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	team, err := createTeam(ctx, client)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = createProject(ctx, client)
	if err != nil {
		log.Fatal(err.Error())
	}

	key, err := client.Team.GenerateAPIKey(ctx, team.UUID)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = fmt.Print(key.Key)
	if err != nil {
		log.Fatal(err.Error())
	}
}
