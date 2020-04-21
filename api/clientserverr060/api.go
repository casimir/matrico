// This file is generated and not meant to be edited by hand."

package clientserverr060

import (
	"net/http"
	"net/url"

	"github.com/casimir/matrico/api/common"
	"github.com/go-chi/chi"
)

func RegisterAPI(r chi.Router) {
	r.Get("/_matrix/client/r0/login", GetLoginFlows)
	r.Post("/_matrix/client/r0/login", Login)
	r.Post("/_matrix/client/r0/register", Register)
	r.Get("/_matrix/client/versions", GetVersions)
	r.Route("/", func(r chi.Router) {
		r.Use(common.AuthorizationMiddleware)
		r.Post("/_matrix/client/r0/user/{userId}/filter", DefineFilter)
		r.Post("/_matrix/client/r0/logout", Logout)
		r.Put("/_matrix/client/r0/presence/{userId}/status", SetPresence)
		r.Get("/_matrix/client/r0/presence/{userId}/status", GetPresence)
		r.Get("/_matrix/client/r0/pushrules/", GetPushRules)
		r.Get("/_matrix/client/r0/sync", Sync)
		r.Get("/_matrix/client/r0/account/whoami", GetTokenOwner)
	})
}

type DefineFilterBody map[string]interface{}

type DefineFilterResponse struct {
	// The ID of the filter that was created. Cannot start
	// with a ``{`` as this character is used to determine
	// if the filter provided is inline JSON or a previously
	// declared filter by homeservers on some APIs.
	FilterID string `json:"filter_id"`
}

// Uploads a new filter definition to the homeserver.
// Returns a filter ID that may be used in future requests to
// restrict which events are returned to the client.
func DefineFilter(w http.ResponseWriter, r *http.Request) {
	userId, erruserId := url.QueryUnescape(chi.URLParam(r, "userId"))
	if erruserId != nil {
		common.ResponseHandler(w, nil, erruserId)
		return
	}
	var body DefineFilterBody
	if err := common.UnmarshalBody(r, &body); err != nil {
		common.ResponseHandler(w, nil, err)
		return
	}
	data, err := defineFilter(r.Context(), userId, body)
	common.ResponseHandler(w, data, err)
}

type GetLoginFlowsResponseFlows struct {
	// The login type. This is supplied as the ``type`` when
	// logging in.
	Type string `json:"type,omitempty"`
}

type GetLoginFlowsResponse struct {
	// The homeserver's supported login types
	Flows []GetLoginFlowsResponseFlows `json:"flows,omitempty"`
}

// Gets the homeserver's supported login types to authenticate users. Clients
// should pick one of these and supply it as the ``type`` when logging in.
func GetLoginFlows(w http.ResponseWriter, r *http.Request) {
	data, err := getLoginFlows(r.Context())
	common.ResponseHandler(w, data, err)
}

type LoginBody struct {
	// Third party identifier for the user.  Deprecated in favour of ``identifier``.
	Address string `json:"address,omitempty"`
	// ID of the client device. If this does not correspond to a
	// known client device, a new device will be created. The server
	// will auto-generate a device_id if this is not specified.
	DeviceID string `json:"device_id,omitempty"`
	// Identification information for the user.
	Identifier map[string]interface{} `json:"identifier,omitempty"`
	// A display name to assign to the newly-created device. Ignored
	// if ``device_id`` corresponds to a known device.
	InitialDeviceDisplayName string `json:"initial_device_display_name,omitempty"`
	// When logging in using a third party identifier, the medium of the identifier. Must be 'email'.  Deprecated in favour of ``identifier``.
	Medium string `json:"medium,omitempty"`
	// Required when ``type`` is ``m.login.password``. The user's
	// password.
	Password string `json:"password,omitempty"`
	// Required when ``type`` is ``m.login.token``. Part of `Token-based`_ login.
	Token string `json:"token,omitempty"`
	// The login type being used.
	Type string `json:"type"`
	// The fully qualified user ID or just local part of the user ID, to log in.  Deprecated in favour of ``identifier``.
	User string `json:"user,omitempty"`
}

type LoginResponse struct {
	// An access token for the account.
	// This access token can then be used to authorize other requests.
	AccessToken string `json:"access_token,omitempty"`
	// ID of the logged-in device. Will be the same as the
	// corresponding parameter in the request, if one was specified.
	DeviceID string `json:"device_id,omitempty"`
	// The server_name of the homeserver on which the account has
	// been registered.
	//
	// **Deprecated**. Clients should extract the server_name from
	// ``user_id`` (by splitting at the first colon) if they require
	// it. Note also that ``homeserver`` is not spelt this way.
	HomeServer string `json:"home_server,omitempty"`
	// The fully-qualified Matrix ID that has been registered.
	UserID string `json:"user_id,omitempty"`
	// Optional client configuration provided by the server. If present,
	// clients SHOULD use the provided object to reconfigure themselves,
	// optionally validating the URLs within. This object takes the same
	// form as the one returned from .well-known autodiscovery.
	WellKnown map[string]interface{} `json:"well_known,omitempty"`
}

// Authenticates the user, and issues an access token they can
// use to authorize themself in subsequent requests.
//
// If the client does not supply a ``device_id``, the server must
// auto-generate one.
//
// The returned access token must be associated with the ``device_id``
// supplied by the client or generated by the server. The server may
// invalidate any access token previously associated with that device. See
// `Relationship between access tokens and devices`_.
func Login(w http.ResponseWriter, r *http.Request) {
	var body LoginBody
	if err := common.UnmarshalBody(r, &body); err != nil {
		common.ResponseHandler(w, nil, err)
		return
	}
	data, err := login(r.Context(), body)
	common.ResponseHandler(w, data, err)
}

type LogoutResponse map[string]interface{}

// Invalidates an existing access token, so that it can no longer be used for
// authorization. The device associated with the access token is also deleted.
// `Device keys <#device-keys>`_ for the device are deleted alongside the device.
func Logout(w http.ResponseWriter, r *http.Request) {
	data, err := logout(r.Context())
	common.ResponseHandler(w, data, err)
}

type SetPresenceBody struct {
	// The new presence state.
	Presence string `json:"presence"`
	// The status message to attach to this state.
	StatusMsg string `json:"status_msg,omitempty"`
}

type SetPresenceResponse map[string]interface{}

// This API sets the given user's presence state. When setting the status,
// the activity time is updated to reflect that activity; the client does
// not need to specify the ``last_active_ago`` field. You cannot set the
// presence state of another user.
func SetPresence(w http.ResponseWriter, r *http.Request) {
	userId, erruserId := url.QueryUnescape(chi.URLParam(r, "userId"))
	if erruserId != nil {
		common.ResponseHandler(w, nil, erruserId)
		return
	}
	var body SetPresenceBody
	if err := common.UnmarshalBody(r, &body); err != nil {
		common.ResponseHandler(w, nil, err)
		return
	}
	data, err := setPresence(r.Context(), userId, body)
	common.ResponseHandler(w, data, err)
}

type GetPresenceResponse struct {
	// Whether the user is currently active
	CurrentlyActive bool `json:"currently_active,omitempty"`
	// The length of time in milliseconds since an action was performed
	// by this user.
	LastActiveAgo int `json:"last_active_ago,omitempty"`
	// This user's presence.
	Presence string `json:"presence"`
	// The state message for this user if one was set.
	StatusMsg *string `json:"status_msg,omitempty"`
}

// Get the given user's presence state.
func GetPresence(w http.ResponseWriter, r *http.Request) {
	userId, erruserId := url.QueryUnescape(chi.URLParam(r, "userId"))
	if erruserId != nil {
		common.ResponseHandler(w, nil, erruserId)
		return
	}
	data, err := getPresence(r.Context(), userId)
	common.ResponseHandler(w, data, err)
}

type GetPushRulesResponseGlobal map[string]interface{}

type GetPushRulesResponse struct {
	// The global ruleset.
	Global GetPushRulesResponseGlobal `json:"global"`
}

// Retrieve all push rulesets for this user. Clients can "drill-down" on
// the rulesets by suffixing a ``scope`` to this path e.g.
// ``/pushrules/global/``. This will return a subset of this data under the
// specified key e.g. the ``global`` key.
func GetPushRules(w http.ResponseWriter, r *http.Request) {
	data, err := getPushRules(r.Context())
	common.ResponseHandler(w, data, err)
}

type RegisterBody struct {
	// Additional authentication information for the
	// user-interactive authentication API. Note that this
	// information is *not* used to define how the registered user
	// should be authenticated, but is instead used to
	// authenticate the ``register`` call itself.
	Auth map[string]interface{} `json:"auth,omitempty"`
	// ID of the client device. If this does not correspond to a
	// known client device, a new device will be created. The server
	// will auto-generate a device_id if this is not specified.
	DeviceID string `json:"device_id,omitempty"`
	// If true, an ``access_token`` and ``device_id`` should not be
	// returned from this call, therefore preventing an automatic
	// login. Defaults to false.
	InhibitLogin bool `json:"inhibit_login,omitempty"`
	// A display name to assign to the newly-created device. Ignored
	// if ``device_id`` corresponds to a known device.
	InitialDeviceDisplayName string `json:"initial_device_display_name,omitempty"`
	// The desired password for the account.
	Password string `json:"password,omitempty"`
	// The basis for the localpart of the desired Matrix ID. If omitted,
	// the homeserver MUST generate a Matrix ID local part.
	Username string `json:"username,omitempty"`
}

type RegisterResponse struct {
	// An access token for the account.
	// This access token can then be used to authorize other requests.
	// Required if the ``inhibit_login`` option is false.
	AccessToken string `json:"access_token,omitempty"`
	// ID of the registered device. Will be the same as the
	// corresponding parameter in the request, if one was specified.
	// Required if the ``inhibit_login`` option is false.
	DeviceID string `json:"device_id,omitempty"`
	// The server_name of the homeserver on which the account has
	// been registered.
	//
	// **Deprecated**. Clients should extract the server_name from
	// ``user_id`` (by splitting at the first colon) if they require
	// it. Note also that ``homeserver`` is not spelt this way.
	HomeServer string `json:"home_server,omitempty"`
	// The fully-qualified Matrix user ID (MXID) that has been registered.
	//
	// Any user ID returned by this API must conform to the grammar given in the
	// `Matrix specification <../appendices.html#user-identifiers>`_.
	UserID string `json:"user_id"`
}

// This API endpoint uses the `User-Interactive Authentication API`_, except in
// the cases where a guest account is being registered.
//
// Register for an account on this homeserver.
//
// There are two kinds of user account:
//
// - `user` accounts. These accounts may use the full API described in this specification.
//
// - `guest` accounts. These accounts may have limited permissions and may not be supported by all servers.
//
// If registration is successful, this endpoint will issue an access token
// the client can use to authorize itself in subsequent requests.
//
// If the client does not supply a ``device_id``, the server must
// auto-generate one.
//
// The server SHOULD register an account with a User ID based on the
// ``username`` provided, if any. Note that the grammar of Matrix User ID
// localparts is restricted, so the server MUST either map the provided
// ``username`` onto a ``user_id`` in a logical manner, or reject
// ``username``\s which do not comply to the grammar, with
// ``M_INVALID_USERNAME``.
//
// Matrix clients MUST NOT assume that localpart of the registered
// ``user_id`` matches the provided ``username``.
//
// The returned access token must be associated with the ``device_id``
// supplied by the client or generated by the server. The server may
// invalidate any access token previously associated with that device. See
// `Relationship between access tokens and devices`_.
//
// When registering a guest account, all parameters in the request body
// with the exception of ``initial_device_display_name`` MUST BE ignored
// by the server. The server MUST pick a ``device_id`` for the account
// regardless of input.
//
// Any user ID returned by this API must conform to the grammar given in the
// `Matrix specification <../appendices.html#user-identifiers>`_.
func Register(w http.ResponseWriter, r *http.Request) {
	var body RegisterBody
	if err := common.UnmarshalBody(r, &body); err != nil {
		common.ResponseHandler(w, nil, err)
		return
	}
	data, err := register(r.Context(), body, r.URL.Query())
	common.ResponseHandler(w, data, err)
}

type SyncResponseAccountData map[string]interface{}

type SyncResponseDeviceLists map[string]interface{}

type SyncResponsePresence map[string]interface{}

type SyncResponseRooms struct {
	// The rooms that the user has been invited to, mapped as room ID to
	// room information.
	Invite map[string]interface{} `json:"invite,omitempty"`
	// The rooms that the user has joined, mapped as room ID to
	// room information.
	Join map[string]interface{} `json:"join,omitempty"`
	// The rooms that the user has left or been banned from, mapped as room ID to
	// room information.
	Leave map[string]interface{} `json:"leave,omitempty"`
}

type SyncResponseToDevice map[string]interface{}

type SyncResponse struct {
	// The global private data created by this user.
	AccountData SyncResponseAccountData `json:"account_data,omitempty"`
	// Information on end-to-end device updates, as specified in
	// |device_lists_sync|_.
	DeviceLists SyncResponseDeviceLists `json:"device_lists,omitempty"`
	// Information on end-to-end encryption keys, as specified
	// in |device_lists_sync|_.
	DeviceOneTimeKeysCount map[string]int `json:"device_one_time_keys_count,omitempty"`
	// The batch token to supply in the ``since`` param of the next
	// ``/sync`` request.
	NextBatch string `json:"next_batch"`
	// The updates to the presence status of other users.
	Presence SyncResponsePresence `json:"presence,omitempty"`
	// Updates to rooms.
	Rooms SyncResponseRooms `json:"rooms,omitempty"`
	// Information on the send-to-device messages for the client
	// device, as defined in |send_to_device_sync|_.
	ToDevice SyncResponseToDevice `json:"to_device,omitempty"`
}

// Synchronise the client's state with the latest state on the server.
// Clients use this API when they first log in to get an initial snapshot
// of the state on the server, and then continue to call this API to get
// incremental deltas to the state, and to receive new messages.
//
// *Note*: This endpoint supports lazy-loading. See `Filtering <#filtering>`_
// for more information. Lazy-loading members is only supported on a ``StateFilter``
// for this endpoint. When lazy-loading is enabled, servers MUST include the
// syncing user's own membership event when they join a room, or when the
// full state of rooms is requested, to aid discovering the user's avatar &
// displayname.
//
// Like other members, the user's own membership event is eligible
// for being considered redundant by the server. When a sync is ``limited``,
// the server MUST return membership events for events in the gap
// (between ``since`` and the start of the returned timeline), regardless
// as to whether or not they are redundant.  This ensures that joins/leaves
// and profile changes which occur during the gap are not lost.
func Sync(w http.ResponseWriter, r *http.Request) {
	data, err := sync(r.Context(), r.URL.Query())
	common.ResponseHandler(w, data, err)
}

type GetVersionsResponse struct {
	// Experimental features the server supports. Features not listed here,
	// or the lack of this property all together, indicate that a feature is
	// not supported.
	UnstableFeatures map[string]bool `json:"unstable_features,omitempty"`
	// The supported versions.
	Versions []string `json:"versions"`
}

// Gets the versions of the specification supported by the server.
//
// Values will take the form ``rX.Y.Z``.
//
// Only the latest ``Z`` value will be reported for each supported ``X.Y`` value.
// i.e. if the server implements ``r0.0.0``, ``r0.0.1``, and ``r1.2.0``, it will report ``r0.0.1`` and ``r1.2.0``.
//
// The server may additionally advertise experimental features it supports
// through ``unstable_features``. These features should be namespaced and
// may optionally include version information within their name if desired.
// Features listed here are not for optionally toggling parts of the Matrix
// specification and should only be used to advertise support for a feature
// which has not yet landed in the spec. For example, a feature currently
// undergoing the proposal process may appear here and eventually be taken
// off this list once the feature lands in the spec and the server deems it
// reasonable to do so. Servers may wish to keep advertising features here
// after they've been released into the spec to give clients a chance to
// upgrade appropriately. Additionally, clients should avoid using unstable
// features in their stable releases.
func GetVersions(w http.ResponseWriter, r *http.Request) {
	data, err := getVersions(r.Context())
	common.ResponseHandler(w, data, err)
}

type GetTokenOwnerResponse struct {
	// The user id that owns the access token.
	UserID string `json:"user_id"`
}

// Gets information about the owner of a given access token.
//
// Note that, as with the rest of the Client-Server API,
// Application Services may masquerade as users within their
// namespace by giving a ``user_id`` query parameter. In this
// situation, the server should verify that the given ``user_id``
// is registered by the appservice, and return it in the response
// body.
func GetTokenOwner(w http.ResponseWriter, r *http.Request) {
	data, err := getTokenOwner(r.Context())
	common.ResponseHandler(w, data, err)
}
