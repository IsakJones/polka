package littletests

import (
	"context"
	"fmt"

	"github.com/IsakJones/polka/api/dbstore"
)

func Dbtest() {
	fmt.Println("Hello")
	db := dbstore.New(context.Background())

	rows, err := db.Conn.Query(
		db.Ctx,
		"select * from banks",
	)
	if err != nil {
		fmt.Printf("21%s\n", err)
	}

	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var n int32
		var s string
		err = rows.Scan(&n, &s)
		if err != nil {
			fmt.Printf("31%s\n", err)
		}
		fmt.Printf("%d, %s\n", n, s)
	}

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		fmt.Printf("38%s", err)
	}

	db.Conn.Close()
}
