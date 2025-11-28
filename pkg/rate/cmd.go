package rate

import (
	"darkroom/pkg/database"
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
)

var StoreRate = &cobra.Command{
	Use:  "store-rate",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		db, err := database.InitClient()
		if err != nil {
			return err
		}
		rates := &rates{
			db: db,
		}
		picRates, err := rates.read(dir)
		if err != nil {
			return err
		}
		picturesWithStats, err := rates.storeAndUpdateStats(picRates)
		if err != nil {
			return err
		}
		jsonResult, err := json.MarshalIndent(picturesWithStats, "", "  ")
		if err != nil {
			return err
		}
		log.Println(string(jsonResult))
		return nil
	},
}

var ResetRate = &cobra.Command{
	Use:  "reset-rate",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		rates := &rates{}
		err := rates.reset(dir)
		if err != nil {
			return err
		}
		return nil
	},
}

var PickRated = &cobra.Command{
	Use:  "pick-rated",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		output := args[1]
		rates := &rates{}
		err := rates.pickRated(dir, output)
		if err != nil {
			return err
		}
		return nil
	},
}
