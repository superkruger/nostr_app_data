package skmongo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// This file will add support to easily write in integration tests with a mongo database.
// Implement as follows:
//  if testtools.Testing() {
//    return
//  }
//  ctx := context.Background()
//  db, testCollection, cleanupCallback := dpgmongo.DatabaseForTest(ctx, t, collection, mongoSecret)
//  repo := &exceptionsRepository{
//    c: db.Collection(testCollection),
//  }
//  t.Cleanup(cleanupCallback)
//
//  tests := map[string]struct {...}
//
//  for name, testCase := range tests {
//    t.Run(name, func(t *testing.T) {
//      dpgmongo.TransactionTest(ctx, t, db, func(sessCtx mongo.SessionContext) {
//           ... Implement test!
//      })
//

const (
	// NoRollbackForTest is the key to see the test_collection with data.
	NoRollbackForTest = "TEST_NO_ROLLBACK"

	defaultMigrationPath     = "../ops/migrations"
	maxMigrationPathSearches = 7
)

// DatabaseForTest connects to the database configured and creates a test collection.
// It returns a link to it, a test collection and a callback to drop the test collection.
func DatabaseForTest(ctx context.Context, t *testing.T, collection, mongoSecret string) (Mongo, string, func()) {
	t.Setenv("AWS_XRAY_SDK_DISABLED", "true")

	mngo, err := NewWithOptions(mongoSecret, optionsForTest())
	if err != nil {
		t.Fatal(err)
	}
	db := mngo.database

	testCollection := testCollectionName(collection)
	if err := db.CreateCollection(ctx, testCollection); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrations(ctx, db, collection, testCollection); err != nil {
		t.Fatal(err)
	}
	return Mongo{db}, testCollection, func() {
		if os.Getenv(NoRollbackForTest) != "" {
			return
		}
		if err := db.Collection(testCollection).Drop(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

// TransactionTest performs a transaction for a test.
func TransactionTest(ctx context.Context, t *testing.T, db Mongo, callback func(sessCtx mongo.SessionContext)) {
	sess, err := db.Client().StartSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.EndSession(ctx)

	cb := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// rollback transaction after test
		defer func() {
			if os.Getenv(NoRollbackForTest) != "" {
				return
			}
			if err := sess.AbortTransaction(sessCtx); err != nil {
				t.Fatal(err)
			}
		}()

		callback(sessCtx)
		return nil, nil
	}

	if _, err = sess.WithTransaction(ctx, cb); err != nil {
		t.Fatal(err)
	}
}

func testCollectionName(collection string) string {
	return fmt.Sprintf("test_%s_%d", collection, time.Now().UnixNano())
}

func findMigrationFiles() (path string, files []fs.FileInfo, err error) {
	path = defaultMigrationPath
	for i := 0; i < maxMigrationPathSearches; i++ {
		files, err = ioutil.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			path = "../" + path
			continue
		}
		return
	}
	return "", nil, nil
}

func applyMigrations(ctx context.Context, db *mongo.Database, srcCollection, testCollection string) error {
	migrationFilesPath, files, err := findMigrationFiles()
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.json") {
			continue
		}
		dat, err := os.ReadFile(path.Join(migrationFilesPath, f.Name()))
		if err != nil {
			return err
		}
		var commands []bson.D
		if err = bson.UnmarshalExtJSON(dat, true, &commands); err != nil {
			return err
		}
		for i := range commands {
			for j := range commands[i] {
				if commands[i][j].Key == "createIndexes" && commands[i][j].Value.(string) == srcCollection {
					commands[i][j].Value = testCollection
					if err := db.RunCommand(ctx, commands[i]).Err(); err != nil {
						return fmt.Errorf("failed to execute command %v :%w", commands[i], err)
					}
				}
			}
		}
	}
	return nil
}
