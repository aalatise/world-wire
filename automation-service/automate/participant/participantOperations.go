package participant

import (
	"github.com/IBM/world-wire/automation-service/constant"
	"github.com/IBM/world-wire/automation-service/model/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Shim model for ingesting the MongoDB data for instituitions
type instituition struct {
	Nodes []model.Automation `bson:"nodes"`
}

func (op DeploymentOperations) checkParticipantRecord(input model.Automation) ([]string, error) {
	collection, ctx := op.Session.GetCollection()

	var result instituition
	objectID, err := primitive.ObjectIDFromHex(*input.InstitutionID)

	if err := collection.FindOne(
		ctx,
		bson.M{
			"_id":                 objectID,
			"nodes.participantId": *input.ParticipantID,
		},
		options.FindOne().SetProjection(bson.M{
			"nodes.status": 1,
			"_id":          0,
			"nodes.$":      1,
		}),
	).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			LOGGER.Warning("Unable to find participant with ID %v", *input.ParticipantID)
			// // Create participant's record in the institution table if it doesn't exist
			// LOGGER.Infof("Creating participant record in institution table, pid: %v, iid: %v", *input.ParticipantID, *input.InstitutionID)
			// err = op.createParticipantRecord(input)
			// if err != nil {
			return nil, err
			// }
			// return nil, nil
		}
		LOGGER.Errorf("Error in checkParticipantRecord, Error:%v", err)
		return nil, err
	}
	// Participant record exists and we need to get the statuses
	var statusToReturn []string

	// Dive through the nested field for the main status which is always first indexed in the status array
	mainStatus := result.Nodes[0].Status[0]
	statuses := result.Nodes[0].Status
	// If only the main status exists
	if len(statuses) == 1 {
		LOGGER.Debugf("ParticipantID: %v, Status length: 1", *input.ParticipantID)

		if mainStatus == constant.StatusConfiguring {
			LOGGER.Debugf("ParticipantID: %v, status is: configuring", *input.ParticipantID)
			// Set initialized to true
			update := bson.M{
				"$set": bson.M{
					"nodes.$.initialized": true,
				},
			}
			opts := options.FindOneAndUpdate().SetUpsert(false)
			collection, ctx := op.Session.GetCollection()
			filter := bson.M{
				"_id":                 objectID,
				"nodes.participantId": *input.ParticipantID,
			}
			err = collection.FindOneAndUpdate(ctx, filter, update, opts).Err()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", *input.ParticipantID, *input.InstitutionID)
				} else {
					LOGGER.Errorf("Error in checkParticipantRecord\nError: %v", err)
					return nil, err
				}
			}
			return nil, nil
		} else if mainStatus == constant.StatusComplete {
			LOGGER.Debugf("ParticipantID: %v, status is: complete", *input.ParticipantID)
			statusToReturn = append(statusToReturn, constant.StatusComplete)
			return statusToReturn, nil
		} else if mainStatus == constant.StatusPending {
			LOGGER.Debugf("ParticipantID: %v, status is: pending", *input.ParticipantID)
			// Switch particiapnt's status from pending to configuring
			update := bson.M{
				"$set": bson.M{
					"nodes.$.status":      []string{constant.StatusConfiguring},
					"nodes.$.initialized": true,
				},
			}
			opts := options.FindOneAndUpdate().SetUpsert(false)
			collection, ctx := op.Session.GetCollection()
			filter := bson.M{
				"_id":                 objectID,
				"nodes.participantId": *input.ParticipantID,
			}
			// statusToReturn = append(statusToReturn, constant.StatusConfiguring)
			err = collection.FindOneAndUpdate(ctx, filter, update, opts).Err()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", *input.ParticipantID, *input.InstitutionID)
				} else {
					LOGGER.Errorf("Error in checkParticipantRecord\nError: %v", err)
					return nil, err
				}
			}
			return nil, nil
		} else if mainStatus == constant.StatusConfigurationFailed {
			LOGGER.Debugf("status is: %s", mainStatus)
			op.updateParticipantStatus(*input.InstitutionID, *input.ParticipantID, constant.StatusConfiguring, "configure", statuses, false)
			statusToReturn = append(statusToReturn, constant.StatusCreatePREntryFailed)
			return statusToReturn, nil
		} else {
			LOGGER.Debugf("status is: %s", mainStatus)
			op.updateParticipantStatus(*input.InstitutionID, *input.ParticipantID, constant.StatusConfiguring, "configure", statuses, false)
			statusToReturn = append(statusToReturn, mainStatus)
			return statusToReturn, nil
		}
	} else {
		LOGGER.Debug("More then one status")
		for _, s := range statuses {
			if s == constant.StatusConfiguring {
				LOGGER.Debugf("ParticipantID: %v, status is: configuring", *input.ParticipantID)
				continue
			} else if s == constant.StatusPending {
				LOGGER.Debugf("ParticipantID: %v, status is: pending", *input.ParticipantID)
				op.updateParticipantStatus(*input.InstitutionID, *input.ParticipantID, constant.StatusConfiguring, "configure", statuses, false)
			} else {
				statusToReturn = append(statusToReturn, s)
			}
		}
	}
	return statusToReturn, nil
}

// func (op DeploymentOperations) createParticipantRecord(input model.Automation) error {
// 	collection, ctx := op.Session.GetCollection()
// 	objectID, err := primitive.ObjectIDFromHex(*input.InstitutionID)
// 	if err != nil {
// 		LOGGER.Errorf("Unable get ObjectID from Hex: %v", *input.InstitutionID)
// 		return err
// 	}

// 	input.Status = []string{constant.StatusConfiguring}
// 	isInitialized := true
// 	input.Initialized = &isInitialized
// 	opts := options.FindOneAndUpdate().SetUpsert(false)
// 	filter := bson.M{"_id": objectID}
// 	update := bson.M{
// 		"$push": bson.M{
// 			"nodes": input,
// 		},
// 	}
// 	var updatedDocument bson.M
// 	err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDocument)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			LOGGER.Errorf("Unable to find instituition with ID: %v", *input.InstitutionID)
// 			return err
// 		}
// 		LOGGER.Panicf("Error in createParticipantRecord\nError: %v", err)
// 		return err
// 	}
// 	LOGGER.Infof("Created participant record with ID: %v, in institution ID: %v", *input.ParticipantID, *input.InstitutionID)
// 	return nil
// }

func (op DeploymentOperations) updateParticipantStatus(iid, pid, newStatus, result string, previousStatuses []string, last bool) error {
	LOGGER.Infof("Updating Participant(%v) status in MongoDB", pid)
	// If there was no previous status provided (maybe this is run as a goroutine) and you don't want the call to retrieve previous statuses to be blocking
	if previousStatuses == nil {
		var err error
		previousStatuses, err = op.getParticipantStatus(iid, pid)
		if err != nil {
			LOGGER.Errorf("Error in updateParticipantStatus:\nError: %v", err)
			return err
		}
	}

	collection, ctx := op.Session.GetCollection()
	objectID, err := primitive.ObjectIDFromHex(iid)
	filter := bson.M{
		"_id":                 objectID,
		"nodes.participantId": pid,
	}

	statusExists := false
	// Check if the new status is already in MongoDB
	for _, s := range previousStatuses {
		// The new status exists in MongoDB record
		if s == newStatus {
			if result == "failed" {
				statusExists = true
			} else {
				// Remove these statuses from the participant
				// The result should be done or other types of results
				opts := options.FindOneAndUpdate().SetUpsert(false)
				update := bson.M{
					"$pull": bson.M{
						"nodes.$.status": s,
					},
				}
				err = collection.FindOneAndUpdate(ctx, filter, update, opts).Err()
				if err != nil {
					if err == mongo.ErrNoDocuments {
						LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", pid, iid)
					} else {
						LOGGER.Errorf("Error in updateParticipantStatus\nError: %v", err)
						return err
					}
				}
			}
		} else {
			if s == constant.StatusPending ||
				s == constant.StatusConfiguring ||
				s == constant.StatusComplete ||
				s == constant.StatusConfigurationFailed {
				// Remove these statuses from the participant
				opts := options.FindOneAndUpdate().SetUpsert(false)
				update := bson.M{
					"$pull": bson.M{
						"nodes.$.status": s,
					},
				}
				err = collection.FindOneAndUpdate(ctx, filter, update, opts).Err()
				if err != nil {
					if err == mongo.ErrNoDocuments {
						LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", pid, iid)
					} else {
						LOGGER.Errorf("Error in updateParticipantStatus\nError: %v", err)
						return err
					}
				}
			}
		}
	}

	// A new status and it is not "complete"
	if !statusExists {
		var update bson.M
		opts := options.FindOneAndUpdate().SetUpsert(false)

		if result == "failed" {
			// if the parameter last is true then need to prepend the status `ready`, means no further failing status should be append
			if last {
				update = bson.M{
					"$push": bson.M{
						"nodes.$.status": bson.M{
							"$position": 0,
							"$each":     []string{constant.StatusReady},
						},
					},
				}

				err = collection.FindOneAndUpdate(ctx, filter, update, opts).Err()
				if err != nil {
					if err == mongo.ErrNoDocuments {
						LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", pid, iid)
					} else {
						LOGGER.Errorf("Error in updateParticipantStatus\nError: %v", err)
						return err
					}
				}
			}

			update = bson.M{
				"$push": bson.M{
					"nodes.$.status": newStatus,
				},
			}
		} else if result == "configure" || result == "resolve" {
			// Prepend "configuring"
			update = bson.M{
				"$push": bson.M{
					"nodes.$.status": bson.M{
						"$position": 0,
						"$each":     []string{constant.StatusConfiguring},
					},
				},
			}
		} else if (len(previousStatuses) == 2 && previousStatuses[0] == constant.StatusConfiguring) || (len(previousStatuses) == 1 && previousStatuses[0] == constant.StatusConfiguring) {
			if result == "resolve" {
				update = bson.M{
					"$push": bson.M{
						"nodes.$.status": constant.StatusConfiguring,
					},
				}
			} else {
				update = bson.M{
					"$push": bson.M{
						"nodes.$.status": constant.StatusComplete,
					},
				}
			}
		}
		// If there is no new status to push, then don't do anything
		// If any of the statuses change, then push them statuses here
		if update != nil {
			err = collection.FindOneAndUpdate(ctx, filter, update, opts).Err()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", pid, iid)
				} else {
					LOGGER.Errorf("Error in updateParticipantStatus\nError: %v", err)
					return err
				}
			}
		}
	}

	opts := options.Update().SetUpsert(false)
	update := bson.M{"$set": bson.M{"nodes.$.initialized": true}}
	// Filter is at the top (universal for all queries in this function)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			LOGGER.Warningf("Cannot find record with pid: %v and iid: %v", pid, iid)
		} else {
			LOGGER.Errorf("Error in updateParticipantStatus\nError: %v", err)
			return err
		}
	}

	if res.MatchedCount != 0 {
		LOGGER.Debug("Matched and replaced an existing document")
	}

	return err
}

func (op DeploymentOperations) getParticipantStatus(iid, pid string) ([]string, error) {
	collection, ctx := op.Session.GetCollection()
	objectID, err := primitive.ObjectIDFromHex(iid)
	var statusToReturn []string

	if err != nil {
		LOGGER.Warning("Unable to get object ID from iid")
	}

	var result instituition
	if err := collection.FindOne(
		ctx,
		bson.M{
			"nodes.participantId": pid,
			"_id":                 objectID,
		},
		options.FindOne().SetProjection(bson.M{
			"nodes.status": 1,
			"_id":          0,
		}),
	).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			LOGGER.Warning("Unable to find participant with ID %v", pid)
			return nil, err
		}
		LOGGER.Errorf("Error in updateParticipantStatus\nError: %v", err)
		return nil, err
	}
	statusToReturn = result.Nodes[0].Status
	return statusToReturn, nil
}
