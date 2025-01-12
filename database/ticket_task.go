package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/google/uuid"

	"github.com/SecurityBrewery/catalyst/database/busdb"
	"github.com/SecurityBrewery/catalyst/generated/models"
)

func (db *Database) TaskGet(ctx context.Context, id int64, playbookID string, taskID string) (*models.TicketWithTickets, *models.PlaybookResponse, *models.TaskWithContext, error) {
	inc, err := db.TicketGet(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}

	playbook, ok := inc.Playbooks[playbookID]
	if !ok {
		return nil, nil, nil, errors.New("playbook does not exist")
	}

	task, ok := playbook.Tasks[taskID]
	if !ok {
		return nil, nil, nil, errors.New("task does not exist")
	}

	return inc, playbook, &models.TaskWithContext{
		PlaybookId:   playbookID,
		PlaybookName: playbook.Name,
		TaskId:       taskID,
		Task:         *task,
		TicketId:     id,
		TicketName:   inc.Name,
	}, nil
}

func (db *Database) TaskComplete(ctx context.Context, id int64, playbookID string, taskID string, data interface{}) (*models.TicketWithTickets, error) {
	inc, err := db.TicketGet(ctx, id)
	if err != nil {
		return nil, err
	}

	completable := inc.Playbooks[playbookID].Tasks[taskID].Active
	if !completable {
		return nil, errors.New("cannot be completed")
	}

	ticketFilterQuery, ticketFilterVars, err := db.Hooks.TicketWriteFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `LET d = DOCUMENT(@@collection, @ID)
	` + ticketFilterQuery + `
	LET playbook = d.playbooks[@playbookID]
	LET task = playbook.tasks[@taskID]
	LET newtask = MERGE(task, {"data": NOT_NULL(@data, {}), "done": true, closed: @closed })
	LET newtasks = MERGE(playbook.tasks, { @taskID: newtask } )
	LET newplaybook = MERGE(playbook, {"tasks": newtasks})
	LET newplaybooks = MERGE(d.playbooks, { @playbookID: newplaybook } )
	
	UPDATE d WITH { "modified": DATE_ISO8601(DATE_NOW()), "playbooks": newplaybooks } IN @@collection
	RETURN NEW`
	ticket, err := db.ticketGetQuery(ctx, id, query, mergeMaps(map[string]interface{}{
		"playbookID": playbookID,
		"taskID":     taskID,
		"data":       data,
		"closed":     time.Now().UTC(),
	}, ticketFilterVars), &busdb.Operation{
		OperationType: busdb.Update,
		Ids: []driver.DocumentID{
			driver.NewDocumentID(TicketCollectionName, fmt.Sprintf("%d", id)),
		},
		Msg: fmt.Sprintf("Completed task %s in playbook %s", taskID, playbookID),
	})
	if err != nil {
		return nil, err
	}

	playbook := ticket.Playbooks[playbookID]
	task := playbook.Tasks[taskID]

	runNextTasks(id, playbookID, task.Next, task.Data, extractTicketResponse(ticket), db)

	return ticket, nil
}

func extractTicketResponse(ticket *models.TicketWithTickets) *models.TicketResponse {
	return &models.TicketResponse{
		Artifacts:  ticket.Artifacts,
		Comments:   ticket.Comments,
		Created:    ticket.Created,
		Details:    ticket.Details,
		Files:      ticket.Files,
		ID:         ticket.ID,
		Modified:   ticket.Modified,
		Name:       ticket.Name,
		Owner:      ticket.Owner,
		Playbooks:  ticket.Playbooks,
		Read:       ticket.Read,
		References: ticket.References,
		Schema:     ticket.Schema,
		Status:     ticket.Status,
		Type:       ticket.Type,
		Write:      ticket.Write,
	}
}

func (db *Database) TaskUpdate(ctx context.Context, id int64, playbookID string, taskID string, task *models.Task) (*models.TicketWithTickets, error) {
	ticketFilterQuery, ticketFilterVars, err := db.Hooks.TicketWriteFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `LET d = DOCUMENT(@@collection, @ID)
	` + ticketFilterQuery + `
	LET playbook = d.playbooks[@playbookID]
	LET newtasks = MERGE(playbook.tasks, { @taskID: @task } )
	LET newplaybook = MERGE(playbook, {"tasks": newtasks})
	LET newplaybooks = MERGE(d.playbooks, { @playbookID: newplaybook } )
	
	UPDATE d WITH { "modified": DATE_ISO8601(DATE_NOW()), "playbooks": newplaybooks } IN @@collection
	RETURN NEW`
	ticket, err := db.ticketGetQuery(ctx, id, query, mergeMaps(map[string]interface{}{
		"playbookID": playbookID,
		"taskID":     taskID,
		"task":       task,
	}, ticketFilterVars), &busdb.Operation{
		OperationType: busdb.Update,
		Ids: []driver.DocumentID{
			driver.NewDocumentID(TicketCollectionName, fmt.Sprintf("%d", id)),
		},
		Msg: fmt.Sprintf("Saved task %s in playbook %s", taskID, playbookID),
	})
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (db *Database) TaskRun(ctx context.Context, id int64, playbookID string, taskID string) error {
	ticket, _, task, err := db.TaskGet(ctx, id, playbookID, taskID)
	if err != nil {
		return err
	}

	if task.Task.Type == models.TaskTypeAutomation {
		if err := runTask(id, playbookID, taskID, &task.Task, extractTicketResponse(ticket), db); err != nil {
			return err
		}
	}

	return nil
}

func runNextTasks(id int64, playbookID string, next map[string]string, data interface{}, ticket *models.TicketResponse, db *Database) {
	for nextTaskID, requirement := range next {
		nextTask := ticket.Playbooks[playbookID].Tasks[nextTaskID]
		if nextTask.Type == models.TaskTypeAutomation {
			b, err := evalRequirement(requirement, data)
			if err != nil {
				continue
			}
			if b {
				if err := runTask(id, playbookID, nextTaskID, nextTask, ticket, db); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func runTask(ticketID int64, playbookID string, taskID string, task *models.TaskResponse, ticket *models.TicketResponse, db *Database) error {
	playbook := ticket.Playbooks[playbookID]
	msgContext := &models.Context{Playbook: playbook, Task: task, Ticket: ticket}
	origin := &models.Origin{TaskOrigin: &models.TaskOrigin{TaskId: taskID, PlaybookId: playbookID, TicketId: ticketID}}
	jobID := uuid.NewString()
	return publishJobMapping(jobID, *task.Automation, msgContext, origin, task.Payload, db)
}
