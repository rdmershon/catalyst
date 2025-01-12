package automation

import (
	"log"

	"github.com/SecurityBrewery/catalyst/bus"
	"github.com/SecurityBrewery/catalyst/database"
	"github.com/SecurityBrewery/catalyst/generated/models"
)

func resultAutomation(catalystBus *bus.Bus, db *database.Database) error {
	return catalystBus.SubscribeResult(func(resultMsg *bus.ResultMsg) {
		if resultMsg.Target != nil {
			ctx := jobContext()
			switch {
			case resultMsg.Target.TaskOrigin != nil:
				if _, err := db.TaskComplete(
					ctx,
					resultMsg.Target.TaskOrigin.TicketId,
					resultMsg.Target.TaskOrigin.PlaybookId,
					resultMsg.Target.TaskOrigin.TaskId,
					resultMsg.Data,
				); err != nil {
					log.Println(err)
				}
			case resultMsg.Target.ArtifactOrigin != nil:
				enrichment := &models.EnrichmentForm{
					Data: resultMsg.Data,
					Name: resultMsg.Automation,
				}
				_, err := db.EnrichArtifact(ctx, resultMsg.Target.ArtifactOrigin.TicketId, resultMsg.Target.ArtifactOrigin.Artifact, enrichment)
				if err != nil {
					log.Println(err)
				}
			}
		}
	})
}
