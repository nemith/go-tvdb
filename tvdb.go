package tvdb

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PipeList type representing pipe-separated string values.
type PipeList []string

// UnmarshalXML unmarshals an XML element with string value into a pipe separated list of strings.
func (pipeList *PipeList) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	content := ""
	if err := decoder.DecodeElement(&content, &start); err != nil {
		return err
	}

	*pipeList = strings.Split(strings.Trim(content, "|"), "|")
	return nil
}

// Episode represents a TV show episode on TheClient.
type Episode struct {
	ID                    int      `xml:"id"`
	CombinedEpisodeNumber string   `xml:"Combined_episodenumber"`
	CombinedSeason        int      `xml:"Combined_season"`
	DvdChapter            string   `xml:"DVD_chapter"`
	DvdDiscID             string   `xml:"DVD_discid"`
	DvdEpisodeNumber      string   `xml:"DVD_episodenumber"`
	DvdSeason             string   `xml:"DVD_season"`
	Director              PipeList `xml:"Director"`
	EpImgFlag             string   `xml:"EpImgFlag"`
	EpisodeName           string   `xml:"EpisodeName"`
	EpisodeNumber         int      `xml:"EpisodeNumber"`
	FirstAired            string   `xml:"FirstAired"`
	GuestStars            string   `xml:"GuestStars"`
	ImdbID                string   `xml:"IMDB_ID"`
	Language              string   `xml:"Language"`
	Overview              string   `xml:"Overview"`
	ProductionCode        string   `xml:"ProductionCode"`
	Rating                string   `xml:"Rating"`
	RatingCount           string   `xml:"RatingCount"`
	SeasonNumber          int      `xml:"SeasonNumber"`
	Writer                PipeList `xml:"Writer"`
	AbsoluteNumber        string   `xml:"absolute_number"`
	Filename              string   `xml:"filename"`
	LastUpdated           string   `xml:"lastupdated"`
	SeasonID              int      `xml:"seasonid"`
	SeriesID              int      `xml:"seriesid"`
	ThumbAdded            string   `xml:"thumb_added"`
	ThumbHeight           string   `xml:"thumb_height"`
	ThumbWidth            string   `xml:"thumb_width"`
}

type seriesShared struct {
	ID         int    `xml:"id"`
	Language   string `xml:"language"`
	Name       string `xml:"SeriesName"`
	BannerPath string `xml:"banner"`
	Overview   string `xml:"Overview"`
	FirstAired string `xml:"FirstAired"`
	IMDBID     string `xml:"IMDB_ID"`
	Zap2itID   string `xml:"zap2it_id"`
	Network    string `xml:"Network"`
}

// SeriesSummary is returned from GetSeries
type SeriesSummary struct {
	Aliases PipeList `xml:"AliasNames"`
	seriesShared
}

// Series represents TV show on TheClient.
type Series struct {
	Actors        PipeList `xml:"Actors"`
	AirsDayOfWeek string   `xml:"Airs_DayOfWeek"`
	AirsTime      string   `xml:"Airs_Time"`
	ContentRating string   `xml:"ContentRating"`
	Genre         PipeList `xml:"Genre"`
	Network       string   `xml:"Network"`
	Rating        string   `xml:"Rating"`
	RatingCount   string   `xml:"RatingCount"`
	Runtime       string   `xml:"Runtime"`
	Status        string   `xml:"Status"`
	Added         string   `xml:"added"`
	AddedBy       string   `xml:"addedBy"`
	FanartPath    string   `xml:"fanart"`
	LastUpdated   string   `xml:"lastupdated"`
	PostersPath   string   `xml:"posters"`
	seriesShared
}

// Langage format used for Client responses.
type Language struct {
	ID   int    `xml:"id"`
	Abbr string `xml:"abbreviation"`
	Name string `xml:"name"`
}

// Rating of a show or episode for both user rating as well as community
// rating.
type Rating struct {
	ID              int `xml:"id"`
	UserRating      int
	CommunityRating float32
}

// UnmashalXML on Raiting is a hack to combine xml feilds id and seriesid into
// a single field so we can use it for both series and episodes.
func (r *Rating) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	rating := struct {
		ID              int `xml:"id,omitemptu"`
		SeriesID        int `xml:"seriesid,omitempty"`
		UserRating      int
		CommunityRating float32
	}{}
	if err := decoder.DecodeElement(&rating, &start); err != nil {
		return err
	}
	*r = Rating{
		ID:              rating.ID,
		UserRating:      rating.UserRating,
		CommunityRating: rating.CommunityRating,
	}
	if rating.SeriesID != 0 {
		r.ID = rating.SeriesID
	}
	return nil
}

// RemoteSerivce is a supported remote service that can be used by
// SeriesByRemoteID
type RemoteService string

const (
	IMDB   = RemoteService("imdbid")
	Zap2it = RemoteService("zap2it")
)

// Client is the base of all API calls to thetvdb.com.
type Client struct {
	APIKey  string
	BaseURL *url.URL
}

// NewClient returns a new Client API instance.:
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		BaseURL: &url.URL{
			Scheme: "http",
			Host:   "thetvdb.com",
		},
	}
}

// getReponse does the heavy lifting by fetching and decoding API responses.
func getResponse(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed request for '%s' got code '%d'", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	d := xml.NewDecoder(resp.Body)
	if err = d.Decode(v); err != nil {
		return err
	}

	return nil
}

// apiURL returns a base url for the dynamic API with fields already
// populated.
func (c *Client) apiURL(path string, query url.Values) *url.URL {
	u := *c.BaseURL
	u.Path = fmt.Sprintf("api/%s", path)
	u.RawQuery = query.Encode()
	return &u
}

// staticAPIURL returns a base url for the static API with fields already
// populated.
func (c *Client) staticAPIURL(path string) *url.URL {
	u := *c.BaseURL
	u.Path = fmt.Sprintf("api/%s/%s", c.APIKey, path)
	return &u
}

// Lanauges gets a list of lanauges currently supported on Client.
func (c *Client) Languages() ([]Language, error) {
	u := c.staticAPIURL("languages.xml")
	response := struct {
		XMLName xml.Name   `xml:"Languages"`
		Langs   []Language `xml:"Language"`
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}
	return response.Langs, nil
}

// SearchSeries queries for a series by the series name. Returns a slice of
// series summary data.
// See http://thetvdb.com/wiki/index.php?title=API:GetSeries for more information
func (c *Client) SearchSeries(term, lang string) ([]SeriesSummary, error) {
	query := url.Values{}
	query.Set("seriesname", term)
	if lang != "" {
		query.Set("language", lang)
	}

	u := c.apiURL("GetSeries.php", query)

	response := struct {
		XMLName xml.Name `xml:"Data"`
		Series  []SeriesSummary
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}
	return response.Series, nil
}

// SeriesByID gets a single series' details from the Client series id.
func (c *Client) SeriesByID(id int, lang string) (*Series, error) {
	if lang == "" {
		lang = "en"
	}
	u := c.staticAPIURL(fmt.Sprintf("series/%d/%s.xml", id, lang))
	response := struct {
		XMLName xml.Name `xml:"Data"`
		Series  Series
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}

	return &response.Series, nil
}

// SeriesByRemoteID gets a singles series' details from an identifier from a
// remote service like IMDB or Zap2it.
// See: http://thetvdb.com/wiki/index.php?title=API:GetSeriesByRemoteID
func (c *Client) SeriesByRemoteID(service RemoteService, id, lang string) (*SeriesSummary, error) {
	query := url.Values{}
	query.Set(string(service), id)
	if lang != "" {
		query.Set("language", lang)
	}
	u := c.apiURL("GetSeriesByRemoteID.php", query)
	response := struct {
		XMLName xml.Name `xml:"Data"`
		Series  SeriesSummary
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}

	return &response.Series, nil
}

// SeriesAllByID gets a single  series with details as well as a list of all the
// episodes in the series with details.
func (c *Client) SeriesAllByID(id int, lang string) (*Series, []Episode, error) {
	u := c.staticAPIURL(fmt.Sprintf("series/%d/all/%s.xml", id, lang))
	response := struct {
		XMLName  xml.Name `xml:"Data"`
		Series   Series
		Episodes []Episode `xml:"Episode"`
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, nil, err
	}
	return &response.Series, response.Episodes, nil
}

//TODO: Add SeriesEverything to get the zip and parse it
//TODO: Add ActorsBySeries
//TODO: Add BannersBySeries

// EpisodeById gets a single episode by the episode ID.
func (c *Client) EpisodeByID(id int, lang string) (*Episode, error) {
	u := c.staticAPIURL(fmt.Sprintf("episodes/%d/%s.xml", id, lang))
	response := struct {
		XMLName xml.Name `xml:"Data"`
		Episode Episode
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}
	return &response.Episode, nil
}

// episodeBySeries is a common function to get a single episode from a series
// ID, series number, and episode number based on a paticular order such as
// 'dvd' or 'default'
func (c *Client) episodeBySeries(id int, epNum, lang, order string) (*Episode, error) {
	u := c.staticAPIURL(fmt.Sprintf("series/%d/%s/%s/%s.xml", id, order, epNum, lang))
	resp := struct {
		XMLName xml.Name `xml:"Data"`
		Episode Episode
	}{}
	if err := getResponse(u.String(), &resp); err != nil {
		return nil, err
	}
	return &resp.Episode, nil
}

// EpisodeBySeries gets a single episode from the series ID, the season number,
// and the episode number and uses the default series episode numbering.
func (c *Client) EpisodeBySeries(id, season, episode int, lang string) (*Episode, error) {
	epNum := fmt.Sprintf("%d/%d", season, episode)
	return c.episodeBySeries(id, epNum, lang, "default")
}

// EpisodeBySeriesDVD gets a single episode from the series ID, the season number,
// and the episode number and uses the dvd series episode numbering.
func (c *Client) EpisodeBySeriesDVD(id, season, episode int, lang string) (*Episode, error) {
	epNum := fmt.Sprintf("%d/%d", season, episode)
	return c.episodeBySeries(id, epNum, lang, "dvd")
}

// EpisodeBySeriesAbsolute gets a single episode from the series ID, the season number,
// and the episode number and uses the absolute series episode numbering.
func (c *Client) EpisodeBySeriesAbsolute(id, episode int, lang string) (*Episode, error) {
	epNum := fmt.Sprintf("%d", episode)
	return c.episodeBySeries(id, epNum, lang, "absolute")
}

// userFav is the internal function for UserFav, UserFavAdd, and UserFavRemove
// since they all use the same API.
func (c *Client) userFavs(accountID, actionType string, seriesID int) ([]int, error) {
	query := url.Values{}
	query.Set("accountid", accountID)

	if actionType != "" {
		query.Set("type", actionType)
		query.Set("seriesid", strconv.FormatInt(int64(seriesID), 10))
	}

	u := c.apiURL("User_Favorites.php", query)

	data := &struct {
		XMLName xml.Name `xml:"Favorites"`
		Series  []int
	}{}

	if err := getResponse(u.String(), data); err != nil {
		return nil, err
	}
	return data.Series, nil
}

// UserFavs gets a list of a Client's user favorite series.   Returns the series
// IDs.
//
// Note: the accountID here is not the username of the user but rather a special
// accountID.  Users can retrive thier accountIDs from thier user info page @
// http://thetvdb.com/?tab=userinfo.
func (c *Client) UserFavs(accountID string) ([]int, error) {
	return c.userFavs(accountID, "", 0)
}

// UserFavAdd will add a series by the series id to a users favorites. It will
// return the modified list. See UserFavs for information on how to use the
// accountID.
func (c *Client) UserFavAdd(accountID string, seriesID int) ([]int, error) {
	return c.userFavs(accountID, "add", seriesID)
}

// UserFavRemove will delete a series by the series id from the users
// favorites.  It will return the modified list.  See UserFavs for information
// on how to use the accountID.
func (c *Client) UserFavRemove(accountID string, seriesID int) ([]int, error) {
	return c.userFavs(accountID, "remove", seriesID)
}

// ratingResult is used in multiple places so it's it defined as the xml return for
// ratings
type ratingResult struct {
	SerRatings []*Rating `xml:"Series"`
	EpRatings  []*Rating `xml:"Episode"`
}

// userRatings is a common function used for all user rating functions.
func (c *Client) userRatings(accountID string, seriesID int) (*ratingResult, error) {
	query := url.Values{}

	query.Set("apikey", c.APIKey) //Love the consistency of this API
	query.Set("accountid", accountID)
	if seriesID != 0 {
		query.Set("seriesid", strconv.FormatInt(int64(seriesID), 10))
	}
	u := c.apiURL("GetRatingsForUser.php", query)
	result := &ratingResult{}
	if err := getResponse(u.String(), result); err != nil {
		return nil, err
	}

	return result, nil
}

// UserRatings will get the ratings for all series a user has rated.
func (c *Client) UserRatings(accountID string) ([]*Rating, error) {
	result, err := c.userRatings(accountID, 0)
	if err != nil {
		return nil, err
	}

	return result.SerRatings, nil
}

// UserRatingsSeries will get the user raiting for a single series by the
// series ID and return the rating for that series as well as all episodes
// for that series.
func (c *Client) UserRatingsSeries(accountID string, seriesID int) (*Rating, []*Rating, error) {
	result, err := c.userRatings(accountID, seriesID)
	if err != nil {
		return nil, nil, err
	}

	return result.SerRatings[0], result.EpRatings, nil
}

// setUserRating is a common function for both SetUserRatingSeries and
// SetUserRatingEpisode since they utilize the same API.
func (c *Client) setUserRating(accountID, itemType string, itemID, rating int) error {
	if rating < 0 || rating > 10 {
		return fmt.Errorf("Rating must be between 0 and 10 inclusive")
	}

	query := url.Values{}
	query.Set("accountid", accountID)
	query.Set("itemtype", itemType)
	query.Set("itemid", strconv.FormatInt(int64(itemID), 10))
	query.Set("rating", strconv.FormatInt(int64(rating), 10))
	u := c.apiURL("User_Rating.php", query)

	// This API just returns the global rating.  Lets just ignore it
	return getResponse(u.String(), nil)
}

// SetUserRatingSeries will update or set a users rating for a series by series ID
func (c *Client) SetUserRatingSeries(accountID string, seriesID, rating int) error {
	return c.setUserRating(accountID, "series", seriesID, rating)
}

// SetUserRatingEp will update or set a users rating for an episode by episode
// ID.
func (c *Client) SetUserRatingEp(accountID string, epID, rating int) error {
	return c.setUserRating(accountID, "episode", epID, rating)
}

// UserLang will return the prefered language for a user with a given account
// id.
func (c *Client) UserLang(accountID string) (*Language, error) {
	u := c.apiURL("User_PreferredLanguage.php", url.Values{
		"accountid": []string{accountID},
	})

	resp := &struct {
		Lang Language `xml:"Language"`
	}{}
	if err := getResponse(u.String(), resp); err != nil {
		return nil, err
	}

	return &resp.Lang, nil
}
