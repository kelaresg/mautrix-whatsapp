// mautrix-whatsapp - A Matrix-WhatsApp puppeting bridge.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix-whatsapp/database"
	whatsappExt "maunium.net/go/mautrix-whatsapp/whatsapp-ext"
	"strings"
	"time"
)

func (handler *CommandHandler) CommandSpecialMux(ce *CommandEvent) {
	switch ce.Command {
	case "special-create", "special-invite", "special-kick", "special-leave", "special-invite-portal", "special-leave-all":
		if !ce.User.HasSession() {
			ce.Reply("You are not logged in. Use the `login` command to log into WhatsApp.")
			return
		} else if !ce.User.IsConnected() {
			ce.Reply("You are not connected to WhatsApp. Use the `reconnect` command to reconnect.")
			return
		}
		switch ce.Command {
		case "special-create":
			handler.CommandSpecialCreate(ce)
		case "special-invite":
			handler.CommandSpecialInvite(ce)
		case "special-kick":
			handler.CommandSpecialKick(ce)
		case "special-leave":
			handler.CommandSpecialLeave(ce)
		case "special-invite-portal":
			handler.CommandSpecialInvitePortal(ce)
		case "special-leave-all":
			handler.CommandSpecialLeaveAllPortals(ce)
		}
	default:
		ce.Reply("Unknown Command")
	}
}

func (handler *CommandHandler) CommandSpecialHelp(ce *CommandEvent) {
	cmdPrefix := ""
	if ce.User.ManagementRoom != ce.RoomID || ce.User.IsRelaybot {
		cmdPrefix = handler.bridge.Config.Bridge.CommandPrefix + " "
	}

	ce.Reply("* " + strings.Join([]string{
		cmdPrefix + cmdSpecialCreateHelp,
		cmdPrefix + cmdSpecialInviteHelp,
		cmdPrefix + cmdSpecialInvitePortalHelp,
		cmdPrefix + cmdSpecialKickHelp,
		cmdPrefix + cmdSpecialLeaveHelp,
		cmdPrefix + cmdSpecialLeaveAllPortalsHelp,
	}, "\n* "))
}

const cmdSpecialLeaveAllPortalsHelp = `special-leave-all - leave all your bridged portal.`

func (handler *CommandHandler) CommandSpecialLeaveAllPortals(ce *CommandEvent) {
	portals := ce.User.GetPortals()
	leave := func(portal *Portal) {
		if len(portal.MXID) > 0 {
			_, _ = portal.MainIntent().KickUser(portal.MXID, &mautrix.ReqKickUser{
				Reason: "Logout",
				UserID: ce.User.MXID,
			})
		}
	}
	customPuppet := handler.bridge.GetPuppetByCustomMXID(ce.User.MXID)
	if customPuppet != nil && customPuppet.CustomIntent() != nil {
		intent := customPuppet.CustomIntent()
		leave = func(portal *Portal) {
			if len(portal.MXID) > 0 {
				_, _ = intent.LeaveRoom(portal.MXID)
				_, _ = intent.ForgetRoom(portal.MXID)
			}
		}
	}
	for _, portal := range portals {
		leave(portal)
	}
}

const cmdSpecialInviteHelp = `special-invite <_group JID_> <_international phone number_>,... - Invite members to a group.`

func (handler *CommandHandler) CommandSpecialInvite(ce *CommandEvent) {
	if len(ce.Args) < 2 {
		ce.Reply("**Usage:** `special-invite <group JID> <international phone number>,...`")
		return
	}

	user := ce.User
	jid := ce.Args[0]
	userNumbers := strings.Split(ce.Args[1], ",")

	if strings.HasSuffix(jid, whatsappExt.NewUserSuffix) {
		ce.Reply("**Usage:** `invite <group JID> <international phone number>,...`")
		return
	}

	for i, number := range userNumbers {
		userNumbers[i] = number + whatsappExt.NewUserSuffix
	}

	contact, ok := user.Conn.Store.Contacts[jid]
	if !ok {
		ce.Reply("Group JID not found in contacts. Try syncing contacts with `sync` first.")
		return
	}
	handler.log.Debugln("Importing", jid, "for", user)
	portal := user.bridge.GetPortalByJID(database.GroupPortalKey(jid))
	if len(portal.MXID) > 0 {
		portal.Sync(user, contact)
		ce.Reply("Portal room synced.")
	} else {
		portal.Sync(user, contact)
		ce.Reply("Portal room created.")
	}

	handler.log.Debugln("Inviting", userNumbers, "to", jid)
	err := user.Conn.HandleGroupInvite(jid, userNumbers)
	if err != nil {
		ce.Reply("Please confirm that you have permission to invite members.")
	} else {
		ce.Reply("Group invitation sent.\nIf the member fails to join the group, please check your permissions or command parameters")
	}
	time.Sleep(time.Duration(3)*time.Second)
	ce.Reply("Syncing room puppet...")
	chatMap := make(map[string]whatsapp.Chat)
	for _, chat := range user.Conn.Store.Chats {
		if chat.Jid == jid {
			chatMap[chat.Jid]= chat
		}
	}
	user.syncPortals(chatMap, false)
	ce.Reply("Syncing room puppet completed")
}

const cmdSpecialInvitePortalHelp = `special-invite-portal <_international phone number_>,... - Invite members to a group.`

func (handler *CommandHandler) CommandSpecialInvitePortal(ce *CommandEvent) {
	if len(ce.Args) < 1 {
		ce.Reply("**Usage:** `special-invite-portal <international phone number>,...`")
		return
	}
	fmt.Println()
	fmt.Printf("%+v", ce)
	fmt.Println()
	fmt.Println("roomId: ", ce.RoomID)
	user := ce.User
	userNumbers := strings.Split(ce.Args[0], ",")

	portal := handler.bridge.GetPortalByMXID(ce.RoomID)
	if len(portal.Key.JID) < 1 {
		ce.Reply("portal does not exist in the bridge")
		return
	}
	jid := portal.Key.JID
	if strings.HasSuffix(jid, whatsappExt.NewUserSuffix) {
		ce.Reply("**Usage:** `invite-portal <international phone number>,...`")
		return
	}

	for i, number := range userNumbers {
		userNumbers[i] = number + whatsappExt.NewUserSuffix
	}

	handler.log.Debugln("Inviting", userNumbers, "to", jid)
	err := user.Conn.HandleGroupInvite(jid, userNumbers)
	if err != nil {
		ce.Reply("Please confirm that you have permission to invite members.")
	} else {
		ce.Reply("Group invitation sent.\nIf the member fails to join the group, please check your permissions or command parameters")
	}
	time.Sleep(time.Duration(3)*time.Second)
	ce.Reply("Syncing room puppet...")
	chatMap := make(map[string]whatsapp.Chat)
	for _, chat := range user.Conn.Store.Chats {
		if chat.Jid == jid {
			chatMap[chat.Jid]= chat
		}
	}
	user.syncPortals(chatMap, false)
	ce.Reply("Syncing room puppet completed")
}

const cmdSpecialKickHelp = `special-kick <_group JID_> <_international phone number_>,... <_reason_> - Remove members from the group.`

func (handler *CommandHandler) CommandSpecialKick(ce *CommandEvent) {
	if len(ce.Args) < 2 {
		ce.Reply("**Usage:** `special-kick <group JID> <international phone number>,... reason`")
		return
	}

	user := ce.User
	jid := ce.Args[0]
	userNumbers := strings.Split(ce.Args[1], ",")
	reason := "omitempty"
	if len(ce.Args) > 2 {
		reason = ce.Args[0]
	}

	if strings.HasSuffix(jid, whatsappExt.NewUserSuffix) {
		ce.Reply("**Usage:** `kick <group JID> <international phone number>,... reason`")
		return
	}

	contact, ok := user.Conn.Store.Contacts[jid]
	if !ok {
		ce.Reply("Group JID not found in contacts. Try syncing contacts with `sync` first.")
		return
	}
	handler.log.Debugln("Importing", jid, "for", user)
	portal := user.bridge.GetPortalByJID(database.GroupPortalKey(jid))
	if len(portal.MXID) > 0 {
		portal.Sync(user, contact)
		ce.Reply("Portal room synced.")
	} else {
		portal.Sync(user, contact)
		ce.Reply("Portal room created.")
	}

	for i, number := range userNumbers {
		userNumbers[i] = number + whatsappExt.NewUserSuffix
		member := portal.bridge.GetPuppetByJID(number + whatsappExt.NewUserSuffix)
		if member == nil {
			portal.log.Errorln("%s is not a puppet", number)
			return
		}
		_, err := portal.MainIntent().KickUser(portal.MXID, &mautrix.ReqKickUser{
			Reason: reason,
			UserID: member.MXID,
		})
		if err != nil {
			portal.log.Errorln("Error kicking user while command kick:", err)
		}
	}

	handler.log.Debugln("Kicking", userNumbers, "to", jid)
	err := user.Conn.HandleGroupKick(jid, userNumbers)
	if err != nil {
		ce.Reply("Please confirm that you have permission to kick members.")
	} else {
		ce.Reply("Remove operation completed.\nIf the member has not been removed, please check your permissions or command parameters")
	}
}

const cmdSpecialLeaveHelp = `special-leave <_group JID_> - leave a group.`

func (handler *CommandHandler) CommandSpecialLeave(ce *CommandEvent) {
	if len(ce.Args) == 0 {
		ce.Reply("**Usage:** `special-leave <group JID>`")
		return
	}

	user := ce.User
	jid := ce.Args[0]

	if strings.HasSuffix(jid, whatsappExt.NewUserSuffix) {
		ce.Reply("**Usage:** `leave <group JID>`")
		return
	}

	err := user.Conn.HandleGroupLeave(jid)
	if err == nil {
		ce.Reply("Leave operation completed.")
	}

	handler.log.Debugln("Importing", jid, "for", user)
	portal := user.bridge.GetPortalByJID(database.GroupPortalKey(jid))
	if len(portal.MXID) > 0 {
		_, errLeave := portal.MainIntent().LeaveRoom(portal.MXID)
		if errLeave != nil {
			portal.log.Errorln("Error leaving matrix room:", err)
		}
	}
}


const cmdSpecialCreateHelp = `special-create <_subject_> <_international phone number_>,... - Create a group.`

func (handler *CommandHandler) CommandSpecialCreate(ce *CommandEvent) {
	if len(ce.Args) < 2 {
		ce.Reply("**Usage:** `special-create <subject> <international phone number>,...`")
		return
	}

	user := ce.User
	subject := ce.Args[0]
	userNumbers := strings.Split(ce.Args[1], ",")

	for i, number := range userNumbers {
		userNumbers[i] = number + whatsappExt.NewUserSuffix
	}

	handler.log.Debugln("Create Group", subject, "with", userNumbers)
	err := user.Conn.HandleGroupCreate(subject, userNumbers)
	if err != nil {
		ce.Reply("Please confirm that parameters is correct.")
	} else {
		ce.Reply("Syncing group list...")
		time.Sleep(time.Duration(3)*time.Second)
		ce.Reply("Syncing group list completed")
	}
}

//const cmdJoinHelp = `join <_Invitation link|code_> - Join the group via the invitation link.`
//
//func (handler *CommandHandler) CommandJoin(ce *CommandEvent) {
//	if len(ce.Args) == 0 {
//		ce.Reply("**Usage:** `join <Invitation link||code>`")
//		return
//	}
//
//	user := ce.User
//	params := strings.Split(ce.Args[0], "com/")
//
//	jid, err := user.Conn.HandleGroupJoin(params[len(params)-1])
//	if err == nil {
//		ce.Reply("Join operation completed.")
//	}
//
//	contact, ok := user.Conn.Store.Contacts[jid]
//	if !ok {
//		ce.Reply("Group JID not found in contacts. Try syncing contacts with `sync` first.")
//		return
//	}
//	handler.log.Debugln("Importing", jid, "for", user)
//	portal := user.bridge.GetPortalByJID(database.GroupPortalKey(jid))
//	if len(portal.MXID) > 0 {
//		portal.Sync(user, contact)
//		ce.Reply("Portal room synced.")
//	} else {
//		portal.Sync(user, contact)
//		ce.Reply("Portal room created.")
//	}
//}

