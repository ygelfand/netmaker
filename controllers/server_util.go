package controller

import (
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/logic"
	"github.com/gravitl/netmaker/models"
	"github.com/gravitl/netmaker/serverctl"
)

func runServerPeerUpdate() error {
	var serverData = models.ServerUpdateData{
		UpdatePeers: true,
	}
	serverctl.Push(serverData)
	var settings, err = serverctl.Pop()
	if err != nil {
		logger.Log(1, "error during pop,", err.Error())
		return err
	}
	return handlePeerUpdate(&settings.ServerNode)
}

func runServerUpdateIfNeeded(shouldPeersUpdate bool, serverNode models.Node) error {
	// check if a peer/server update is needed
	var serverData = models.ServerUpdateData{
		UpdatePeers: shouldPeersUpdate,
		ServerNode:  serverNode,
	}
	serverctl.Push(serverData)

	return handleServerUpdate()
}

func handleServerUpdate() error {
	var settings, settingsErr = serverctl.Pop()
	if settingsErr != nil {
		return settingsErr
	}
	var currentServerNodeID, err = logic.GetNetworkServerNodeID(settings.ServerNode.Network)
	if err != nil {
		return err
	}
	// ensure server client is available
	if settings.UpdatePeers || (settings.ServerNode.ID == currentServerNodeID) {
		err = serverctl.SyncServerNetwork(&settings.ServerNode)
		if err != nil {
			logger.Log(1, "failed to sync,", settings.ServerNode.Network, ", error:", err.Error())
		}
	}
	// if peers should update, update peers on network
	if settings.UpdatePeers {
		if err = handlePeerUpdate(&settings.ServerNode); err != nil {
			return err
		}
		logger.Log(1, "updated peers on network:", settings.ServerNode.Network)
	}
	// if the server node had an update, run the update function
	if settings.ServerNode.ID == currentServerNodeID {
		if err = logic.ServerUpdate(&settings.ServerNode); err != nil {
			return err
		}
		logger.Log(1, "server node:", settings.ServerNode.ID, "was updated")
	}
	return nil
}

// tells server to update it's peers
func handlePeerUpdate(node *models.Node) error {
	logger.Log(1, "updating peers on network:", node.Network)
	var currentServerNodeID, err = logic.GetNetworkServerNodeID(node.Network)
	if err != nil {
		return err
	}
	var currentServerNode, currErr = logic.GetNodeByID(currentServerNodeID)
	if currErr != nil {
		return currErr
	}
	logic.SetNetworkServerPeers(&currentServerNode)
	logger.Log(1, "finished a peer update for network,", currentServerNode.Network)
	return nil
}