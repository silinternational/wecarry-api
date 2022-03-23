package listeners

import (
	"errors"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5"

	"github.com/silinternational/wecarry-api/cache"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
	"github.com/silinternational/wecarry-api/marketing"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

type listenerContext struct {
	buffalo.DefaultContext
	params map[interface{}]interface{}
}

// Value retrieves a context item added by `Set`
func (lc *listenerContext) Value(key interface{}) interface{} {
	return lc.params[key]
}

// Set a new value on the Context. CAUTION: this is not thread-safe
func (lc *listenerContext) Set(key string, val interface{}) {
	lc.params[key] = val
}

func (*listenerContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*listenerContext) Done() <-chan struct{} {
	return nil
}

func (*listenerContext) Err() error {
	return nil
}

var eventTypes = map[string]func(event events.Event){
	domain.EventApiUserCreated:                    userCreatedHandler,
	domain.EventApiMessageCreated:                 sendNewThreadMessageNotification,
	domain.EventApiRequestStatusUpdated:           sendRequestStatusUpdatedNotification,
	domain.EventApiRequestCreated:                 requestCreatedHandler,
	domain.EventApiRequestUpdated:                 cacheRequestUpdatedListener,
	domain.EventApiPotentialProviderCreated:       potentialProviderCreated,
	domain.EventApiPotentialProviderSelfDestroyed: potentialProviderSelfDestroyed,
	domain.EventApiPotentialProviderRejected:      potentialProviderRejected,
	domain.EventApiMeetingInviteCreated:           meetingInviteCreated,
}

func userCreatedHandler(event events.Event) {
	userCreatedLogger(event)
	userCreatedSendWelcomeMessage(event)
	userCreatedAddToMarketingList(event)
}

func requestCreatedHandler(event events.Event) {
	sendRequestCreatedNotifications(event)
	cacheRequestCreatedListener(event)
}

func listener(e events.Event) {
	defer func() {
		if err := recover(); err != nil {
			domain.ErrLogger.Printf("panic in event %s: %s", e.Kind, err)
		}
	}()

	handler, ok := eventTypes[e.Kind]
	if !ok {
		if strings.HasPrefix(e.Kind, "app") {
			panic("event '" + e.Kind + "' has no handler")
		}
		return
	}

	time.Sleep(time.Second * 5) // a rough guess at the longest time it takes for the database transaction to close

	handler(e)
}

// RegisterListener registers the event listener
func RegisterListener() {
	if _, err := events.Listen(listener); err != nil {
		panic("failed to register event listener " + err.Error())
	}
}

func userCreatedLogger(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	domain.Logger.Printf("User Created: %s", e.Message)
}

func userCreatedSendWelcomeMessage(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	user, ok := e.Payload["user"].(*models.User)
	if !ok {
		domain.ErrLogger.Printf("Failed to get User from event payload for sending welcome message. Event message: %s", e.Message)
		return
	}

	if err := sendNewUserWelcome(*user); err != nil {
		domain.ErrLogger.Printf("Failed to send new user welcome to %s. Error: %s",
			user.UUID.String(), err)
	}
}

func userCreatedAddToMarketingList(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	user, ok := e.Payload["user"].(*models.User)
	if !ok {
		domain.ErrLogger.Printf(
			"Failed to get User from event payload for adding to marketing list. Event message: %s", e.Message)
		return
	}

	// ensure env vars are present
	if domain.Env.MailChimpAPIKey == "" {
		domain.ErrLogger.Printf("missing required env var for MAILCHIMP_API_KEY. need to add %s to list", user.Email)
		return
	}
	if domain.Env.MailChimpListID == "" {
		domain.ErrLogger.Printf("missing required env var for MAILCHIMP_LIST_ID. need to add %s to list", user.Email)
		return
	}
	if domain.Env.MailChimpUsername == "" {
		domain.ErrLogger.Printf("missing required env var for MAILCHIMP_USERNAME. need to add %s to list", user.Email)
		return
	}

	err := marketing.AddUserToList(*user, domain.Env.MailChimpAPIBaseURL, domain.Env.MailChimpListID,
		domain.Env.MailChimpUsername, domain.Env.MailChimpAPIKey)
	if err != nil {
		domain.ErrLogger.Printf("error calling marketing.AddUserToList when trying to add %s: %s",
			user.Email, err.Error())
	}
}

func sendNewThreadMessageNotification(e events.Event) {
	if e.Kind != domain.EventApiMessageCreated {
		return
	}

	domain.Logger.Printf("%s Thread Message Created ... %s", domain.GetCurrentTime(), e.Message)

	id, ok := e.Payload[domain.ArgMessageID].(int)
	if !ok {
		domain.ErrLogger.Printf("sendNewThreadMessageNotification: unable to read message ID from event payload")
		return
	}

	if err := job.SubmitDelayed(job.NewThreadMessage, domain.NewMessageNotificationDelay,
		map[string]interface{}{domain.ArgMessageID: id}); err != nil {
		domain.ErrLogger.Printf("error starting 'New Message' job, %s", err)
	}
}

func sendRequestStatusUpdatedNotification(e events.Event) {
	if e.Kind != domain.EventApiRequestStatusUpdated {
		return
	}

	pEData, ok := e.Payload[domain.ArgEventData].(models.RequestStatusEventData)
	if !ok {
		domain.ErrLogger.Printf("unable to parse Request Status Updated event payload")
		return
	}

	pid := pEData.RequestID

	request := models.Request{}
	if err := request.FindByID(models.DB, pid); err != nil {
		domain.ErrLogger.Printf("unable to find request from event with id %v ... %s", pid, err)
	}

	requestStatusUpdatedNotifications(request, pEData)
}

func sendRequestCreatedNotifications(e events.Event) {
	if e.Kind != domain.EventApiRequestCreated {
		return
	}

	eventData, ok := e.Payload[domain.ArgEventData].(models.RequestCreatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Request Created event payload incorrect type: %T", e.Payload[domain.ArgEventData])
		return
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from request-created event, %s", eventData.RequestID, err)
		return
	}

	users, err := request.GetAudience(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to get request audience in event listener: %s", err.Error())
		return
	}

	sendNewRequestNotifications(request, users)
}

func cacheRequestCreatedListener(e events.Event) {
	if e.Kind != domain.EventApiRequestCreated {
		return
	}

	eventData, ok := e.Payload[domain.ArgEventData].(models.RequestCreatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Request Created event payload incorrect type: %T", e.Payload[domain.ArgEventData])
		return
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from request-created event, %s", eventData.RequestID, err)
	}

	err := models.DB.Transaction(func(tx *pop.Connection) error {
		ctx := newListenerContext()
		ctx.Set(domain.ContextKeyTx, tx)
		return cache.CacheRebuildOnNewRequest(ctx, request)
	})
	if err != nil {
		domain.ErrLogger.Printf("error in cache rebuild on new request: " + err.Error())
	}
}

func cacheRequestUpdatedListener(e events.Event) {
	if e.Kind != domain.EventApiRequestUpdated {
		return
	}

	eventData, ok := e.Payload[domain.ArgEventData].(models.RequestUpdatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Request Updated event payload incorrect type: %T", e.Payload[domain.ArgEventData])
		return
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from request-created event, %s", eventData.RequestID, err)
	}

	err := models.DB.Transaction(func(tx *pop.Connection) error {
		ctx := newListenerContext()
		ctx.Set(domain.ContextKeyTx, tx)
		return cache.CacheRebuildOnChangedRequest(ctx, request)
	})
	if err != nil {
		domain.ErrLogger.Printf("error in cache rebuild on changed request: " + err.Error())
	}
}

func potentialProviderCreated(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderCreated {
		return
	}

	eventData, ok := e.Payload[domain.ArgEventData].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload[domain.ArgEventData])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(models.DB, eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from PotentialProvider event, %s", eventData.RequestID, err)
	}

	creator, err := request.Creator(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to find request %d creator from PotentialProvider event, %s",
			eventData.RequestID, err)
	}

	err = sendPotentialProviderCreatedNotification(potentialProvider.Nickname, creator, request)
	if err != nil {
		domain.ErrLogger.Printf(err.Error())
	}
}

func potentialProviderSelfDestroyed(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderSelfDestroyed {
		return
	}

	eventData, ok := e.Payload[domain.ArgEventData].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload[domain.ArgEventData])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(models.DB, eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from PotentialProvider event, %s", eventData.RequestID, err)
	}

	creator, err := request.Creator(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to find request %d creator from PotentialProvider event, %s",
			eventData.RequestID, err)
	}

	err = sendPotentialProviderSelfDestroyedNotification(potentialProvider.Nickname, creator, request)
	if err != nil {
		domain.ErrLogger.Printf(err.Error())
	}
}

func potentialProviderRejected(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderRejected {
		return
	}

	eventData, ok := e.Payload[domain.ArgEventData].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload[domain.ArgEventData])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(models.DB, eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from PotentialProvider event, %s", eventData.RequestID, err)
	}

	creator, err := request.Creator(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to find request %d creator from PotentialProvider event, %s",
			eventData.RequestID, err)
	}

	err = sendPotentialProviderRejectedNotification(potentialProvider, creator.Nickname, request)
	if err != nil {
		domain.ErrLogger.Printf(err.Error())
	}
}

func sendNewUserWelcome(user models.User) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	language := user.GetLanguagePreference(models.DB)
	subject := domain.GetTranslatedSubject(language, "Email.Subject.Welcome", map[string]string{})

	msg := notifications.Message{
		Template:  domain.MessageTemplateNewUserWelcome,
		ToName:    user.GetRealName(),
		ToEmail:   user.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Subject:   subject,
		Data: map[string]interface{}{
			"appName":      domain.Env.AppName,
			"uiURL":        domain.Env.UIURL,
			"supportEmail": domain.Env.SupportEmail,
			"userEmail":    user.Email,
			"firstName":    user.FirstName,
		},
	}
	return notifications.Send(msg)
}

func meetingInviteCreated(e events.Event) {
	if e.Kind != domain.EventApiMeetingInviteCreated {
		return
	}

	id, err := getID(e.Payload)
	if err != nil {
		domain.ErrLogger.Printf("meeting invite ID not found in payload, %s", err)
		return
	}

	foundMeetingInvite := false
	var invite models.MeetingInvite
	var findErr error

	i := 1
	for ; i <= domain.Env.ListenerMaxRetries; i++ {
		findErr = invite.FindByID(models.DB, id, "Meeting", "Inviter")
		if findErr == nil {
			foundMeetingInvite = true
			break
		}
		time.Sleep(getDelayDuration(i * i))
	}
	domain.Logger.Printf("listener meetingInviteCreated required %d retries with delay %d", i-1, domain.Env.ListenerDelayMilliseconds)

	if !foundMeetingInvite {
		domain.ErrLogger.Printf("failed to find MeetingInvite in meetingInviteCreated, %s", findErr)
		return
	}

	if err = sendMeetingInvite(invite); err != nil {
		domain.ErrLogger.Printf("unable to send invite %d in meetingInviteCreated event, %s", invite.ID, err)
	}
}

// sendMeetingInvite sends an email to the invitee. The MeetingInvite must have its Meeting and Inviter hydrated.
func sendMeetingInvite(invite models.MeetingInvite) error {
	if invite.Email == "" {
		return errors.New("'To' email address is required")
	}

	language := invite.Inviter.GetLanguagePreference(models.DB)
	subject := domain.GetTranslatedSubject(language, "Email.Subject.MeetingInvite",
		map[string]string{"MeetingName": invite.Meeting.Name})

	msg := notifications.Message{
		Template:  domain.MessageTemplateMeetingInvite,
		ToEmail:   invite.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Subject:   subject,
		Data: map[string]interface{}{
			"appName":      domain.Env.AppName,
			"uiURL":        domain.Env.UIURL,
			"supportEmail": domain.Env.SupportEmail,
			"inviterName":  invite.Inviter.FirstName,
			"eventName":    invite.Meeting.Name,
			"inviteURL":    invite.InviteURL(),
		},
	}
	return notifications.Send(msg)
}

// getDelayDuration is a helper function to calculate delay in milliseconds before processing event
func getDelayDuration(multiplier int) time.Duration {
	return time.Duration(domain.Env.ListenerDelayMilliseconds) * time.Millisecond * time.Duration(multiplier)
}

func getID(p events.Payload) (int, error) {
	i, ok := p[domain.ArgId]
	if !ok {
		return 0, errors.New("id not in event payload")
	}

	id, ok := i.(int)
	if !ok {
		return 0, errors.New("ID is not an int")
	}

	return id, nil
}

func newListenerContext() *listenerContext {
	ctx := &listenerContext{
		params: map[interface{}]interface{}{},
	}
	return ctx
}
