// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type AutoRespond interface {
	HandleBusyDM(post *model.Post) error
	SetUserAutoRespondMessage(userID string, message string) error
}

func (m *mscalendar) HandleBusyDM(post *model.Post) error {
	channel, err := m.PluginAPI.GetMattermostChannel(post.ChannelId)
	if err != nil {
		return err
	}

	if channel.Type != model.CHANNEL_DIRECT {
		return nil
	}

	usersInChannel, err := m.PluginAPI.GetMattermostUsersInChannel(post.ChannelId, model.CHANNEL_SORT_BY_USERNAME, 0, 2)
	if err != nil {
		return err
	}

	var storedRecipient *store.User
	for _, u := range usersInChannel {
		storedUser, _ := m.Store.LoadUser(u.Id)
		if u.Id != post.UserId {
			storedRecipient = storedUser
		}
	}

	if storedRecipient == nil {
		return nil
	}

	recipientStatus, err := m.PluginAPI.GetMattermostUserStatus(storedRecipient.MattermostUserID)
	if err != nil {
		return err
	}
	if recipientStatus.Status == model.STATUS_ONLINE {
		return nil
	}

	autoRespond, err := m.Store.GetSetting(storedRecipient.MattermostUserID, store.AutoRespondSettingID)
	if err != nil {
		return err
	}

	autoRespondBool, ok := autoRespond.(bool)
	if !ok {
		return errors.Errorf("Error retrieving setting: %s", store.AutoRespondSettingID)
	}
	if autoRespondBool && len(storedRecipient.ActiveEvents) > 0 {

		autoRespondMessage, err := m.Store.GetSetting(storedRecipient.MattermostUserID, store.AutoRespondMessageSettingID)
		if err != nil {
			return err
		}

		autoRespondMessageString, ok := autoRespondMessage.(string)
		if !ok {
			return errors.Errorf("Error retrieving setting: %s", store.AutoRespondMessageSettingID)
		}
		if autoRespondMessageString == "" {
			autoRespondMessageString = "This user is currently in a meeting."
		}

		m.Poster.Ephemeral(post.UserId, post.ChannelId, autoRespondMessageString)
	}

	return nil
}

func (m *mscalendar) SetUserAutoRespondMessage(userID string, message string) error {
		return m.Store.SetSetting(userID, store.AutoRespondMessageSettingID, message)
}