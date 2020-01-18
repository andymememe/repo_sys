package repodbcontroller

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Package struct
type Package struct {
	Name        string    `bson:"name" json:"name"`
	Version     string    `bson:"version" json:"version"`
	PackageName string    `bson:"package_name" json:"package_name"`
	LastUpdate  time.Time `bson:"last_update" json:"last_update"`
	Status      string    `bson:"status" json:"status"`
}

// RepoDBController defined DB connection for repo
type RepoDBController struct {
	host   string
	user   string
	pwd    string
	port   string
	client *mongo.Client
}

// NewRepoDBController new a RepoDBController
func NewRepoDBController() *RepoDBController {
	return &RepoDBController{
		host: "localhost",
		user: "admin",
		pwd:  "admin",
		port: "27017",
	}
}

// ConnectDB connect to repo DB
func (r *RepoDBController) ConnectDB() error {
	client, err := mongo.NewClient(
		options.
			Client().
			ApplyURI(fmt.Sprintf(
				"mongodb://%s:%s@%s:%s/repo",
				r.user,
				r.pwd,
				r.host,
				r.port,
			)),
	)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}

	r.client = client
	return nil
}

// GetPackagesByName get packages by name, not packages name
func (r *RepoDBController) GetPackagesByName(repoNames []string, name string) ([]Package, error) {
	var err error
	var packages []Package
	var cur *mongo.Cursor

	ctx := context.Background()
	for _, repoName := range repoNames {
		repo := r.client.Database("repo").Collection(repoName)
		opts := options.Find().SetSort(bson.D{
			bson.E{
				Key:   "name",
				Value: 1,
			},
		})
		cur, err = repo.Find(ctx, bson.M{
			"name": bson.M{
				"$regex": ".*" + name + ".*",
			},
		}, opts)
		if err != nil {
			return []Package{}, err
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			elem := &Package{}
			err = cur.Decode(elem)
			if err != nil {
				return packages, err
			}
			packages = append(packages, *elem)
		}

		err = cur.Err()
		if err != nil {
			return []Package{}, err
		}
	}

	return packages, nil
}

// GetPackageByPkgName get packages by package name
func (r *RepoDBController) GetPackageByPkgName(pkgName string, repoName string) (Package, error) {
	var err error
	var pkg Package
	var cur *mongo.Cursor

	ctx := context.Background()
	repo := r.client.Database("repo").Collection(repoName)
	cur, err = repo.Find(ctx, bson.M{
		"package_name": pkgName,
	})
	if err != nil {
		return Package{}, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		pkg = Package{}
		err = cur.Decode(&pkg)
		if err != nil {
			return Package{}, err
		}
	}

	err = cur.Err()
	if err != nil {
		return Package{}, err
	}

	return pkg, nil
}
