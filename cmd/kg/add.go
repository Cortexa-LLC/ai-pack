package kg

import (
	"fmt"
	"github.com/cortexa-llc/ai-pack/internal/knowledge"
	"github.com/spf13/cobra"
)

var addEntityType string
var addEntityName string
var addEntitySummary string

var addEntityCmd = &cobra.Command{
	Use:   "entity",
	Short: "Add a new entity",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := knowledge.OpenStore(".ai/knowledge.db")
		if err != nil {
			return err
		}
		entity, err := store.CreateEntity(addEntityName, addEntityType, "")
		if err != nil {
			return err
		}
		if addEntitySummary != "" {
			obs, err := store.CreateObservation(entity.ID, addEntitySummary, "")
			if err != nil {
				return err
			}
			fmt.Printf("Entity %s created with summary observation: %s\n", entity.ID, obs.ID)
		} else {
			fmt.Printf("Entity %s created.\n", entity.ID)
		}
		return nil
	},
}

var addObservationCmd = &cobra.Command{
	Use:   "observation <entity-id> <content>",
	Short: "Add an observation to an entity",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := knowledge.OpenStore(".ai/knowledge.db")
		if err != nil {
			return err
		}
		obs, err := store.CreateObservation(args[0], args[1], "")
		if err != nil {
			return err
		}
		fmt.Printf("Added observation %s\n", obs.ID)
		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add entity or observation",
}

func init() {
	addEntityCmd.Flags().StringVar(&addEntityType, "type", "concept", "Entity type")
	addEntityCmd.Flags().StringVar(&addEntityName, "name", "", "Entity name")
	addEntityCmd.Flags().StringVar(&addEntitySummary, "summary", "", "Entity summary")
	addEntityCmd.MarkFlagRequired("name")
	addEntityCmd.MarkFlagRequired("type")
	addCmd.AddCommand(addEntityCmd)
	addCmd.AddCommand(addObservationCmd)
	rootCmd.AddCommand(addCmd)
}// AddObservation atomically adds a new observation to an entity
func AddObservation(db *neo4j.Session, entityID string, content string) error {
    // Begin transaction
    tx, err := db.BeginTransaction()
    if err != nil {
        return err
    }
    
    // Execute a single query to create the observation and link it to the entity
    query := "MATCH (e:Entity) WHERE e.id = $entityId CREATE (o:Observation {content: $content}) CREATE (e)-[:HAS_OBSERVATION]->(o)"
    _, err = tx.Run(query, map[string]interface{}{ "entityId": entityID, "content": content })
    if err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit()
}
