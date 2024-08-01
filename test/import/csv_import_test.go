package import_test

import (
	"encoding/csv"
	"os"
	"testing"
	importer "yaba/internal/import"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestValidCSVs(t *testing.T) {
	t.Parallel()
	testCases := []struct{
		filename string
		expectedRows int
	}{
		{
			filename: "all_cols.csv",
			expectedRows: 1,
		},
		{
			filename: "acceptable_dollars.csv",
			expectedRows: 19,
		},
	}
	for _, test := range testCases {
		path := "testdata/" + test.filename
		rows := test.expectedRows
		t.Parallel()
		f, err := os.Open(path)
		require.NoError(t, err)

		owner, err := uuid.NewRandom()
		require.NoError(t, err)

		csvReader := csv.NewReader(f)
		expenditures, err := importer.ImportExpendituresFromCSVReader(owner, csvReader)
		require.NoError(t, err)
		require.Len(t, expenditures, rows)
	}
}
