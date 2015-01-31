package tvdb

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PipeList type representing pipe-separated string values.
type PipeList []string

// UnmarshalXML unmarshals an XML element with string value into a pipe separated list of strings.
func (pipeList *PipeList) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	content := ""
	if err := decoder.DecodeElement(&content, &start); err != nil {
		return err
	}

	// Empty contents mean just use an empty list
	if content != "" {
		*pipeList = strings.Split(strings.Trim(content, "|"), "|")
	} else {
		*pipeList = []string{}
	}
	return nil
}

type ImgFlag int

func (f ImgFlag) IsValid() bool {
	return int(f) == 1 || int(f) == 2
}

func (f *ImgFlag) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var i int
	err := decoder.DecodeElement(&i, &start)

	// Check to see if it's empty and return the zero value
	if nerr, ok := err.(*strconv.NumError); ok && nerr.Num == "" {
		return nil
	} else if err != nil {
		return err
	}

	*f = ImgFlag(i)
	return nil
}

func (f ImgFlag) String() string {
	if s, ok := imgFlagNameMap[f]; ok {
		return s
	}
	return strconv.FormatInt(int64(f), 10)
}

const (
	ImgFlagNone ImgFlag = iota
	ImgFlag4x3
	ImgFlag16x9
	ImgFlagBadAspectRatio
	ImgFlagTooSmall
	ImgFlagBlackBars
	ImgFlagImproperActionShot
)

var imgFlagNameMap = map[ImgFlag]string{
	ImgFlagNone:               "None",
	ImgFlag4x3:                "4:3",
	ImgFlag16x9:               "16x9",
	ImgFlagBadAspectRatio:     "Bad Aspect Ratio",
	ImgFlagTooSmall:           "Image Too Small",
	ImgFlagBlackBars:          "Black Bars",
	ImgFlagImproperActionShot: "Improper Action Shot",
}

type NullInt struct {
	Int   int
	Valid bool
}

func NewNullInt(i int) NullInt {
	return NullInt{i, true}
}

func (i *NullInt) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var j int
	err := decoder.DecodeElement(&j, &start)

	// Check for emptry string
	if nerr, ok := err.(*strconv.NumError); ok && nerr.Num == "" {
		// Returns the zero values which will be 0, false
		return nil
	} else if err != nil {
		return err
	}
	i.Int = j
	i.Valid = true
	// No errors means we parsed the int sucessfully so it is valid
	return nil
}

type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

func NewNullFloat64(f float64) NullFloat64 {
	return NullFloat64{f, true}
}

func (f *NullFloat64) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var j float64
	err := decoder.DecodeElement(&j, &start)

	// Check for emptry string
	if nerr, ok := err.(*strconv.NumError); ok && nerr.Num == "" {
		// Returns the zero values which will be 0, false
		return nil
	} else if err != nil {
		return err
	}
	f.Float64 = j
	f.Valid = true
	// No errors means we parsed the int sucessfully so it is valid
	return nil
}

type UnixEpochTime time.Time

func (t *UnixEpochTime) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var ut int64
	if err := decoder.DecodeElement(&ut, &start); err != nil {
		return err
	}

	*t = UnixEpochTime(time.Unix(ut, int64(0)).UTC())
	return nil
}

type DateTimeStamp time.Time

func NewDateTimeStamp(year int, month time.Month, day, hour, min, sec int) DateTimeStamp {
	return DateTimeStamp(time.Date(year, month, day, hour, min, sec, 0, time.UTC))
}

func (t *DateTimeStamp) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var ts string
	if err := decoder.DecodeElement(&ts, &start); err != nil {
		return err
	}

	if ts == "" {
		*t = NullDateTimeStamp
		return nil
	}

	// Reference Time: Mon Jan 2 15:04:05 -0700 MST 2006
	tm, err := time.Parse("2006-01-02 15:04:05", ts)
	*t = DateTimeStamp(tm)
	return err
}

var (
	NullDateTimeStamp = NewDateTimeStamp(0, time.January, 0, 0, 0, 0)
)

type DateStamp time.Time

func NewDateStamp(year int, month time.Month, day int) DateStamp {
	return DateStamp(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

func (t *DateStamp) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var ts string
	if err := decoder.DecodeElement(&ts, &start); err != nil {
		return err
	}

	if ts == "" {
		// Return nil
		return nil
	}

	// Reference Time: Mon Jan 2 15:04:05 -0700 MST 2006
	tm, err := time.Parse("2006-01-02", ts)
	*t = DateStamp(tm)
	return err
}

// Episode represents a TV show episode on TheTVDB.
type Episode struct {
	ID                    int    `xml:"id"`
	CombinedEpisodeNumber string `xml:"Combined_episodenumber"`
	CombinedSeason        int    `xml:"Combined_season"`
	// TODO: Handle this better.  This cam be 1 or 1.1
	DVDEpisodeNumber string        `xml:"DVD_episodenumber,omitempty"`
	DVDSeason        NullInt       `xml:"DVD_season,omitempty"`
	Director         PipeList      `xml:"Director"`
	EpImgFlag        ImgFlag       `xml:"EpImgFlag"`
	EpisodeName      string        `xml:"EpisodeName"`
	EpisodeNumber    int           `xml:"EpisodeNumber"`
	FirstAired       DateStamp     `xml:"FirstAired"` // Date string
	GuestStars       PipeList      `xml:"GuestStars"`
	ImdbID           string        `xml:"IMDB_ID"`
	Language         string        `xml:"Language"`
	Overview         string        `xml:"Overview"`
	ProductionCode   string        `xml:"ProductionCode"`
	Rating           NullFloat64   `xml:"Rating"`
	RatingCount      NullInt       `xml:"RatingCount"`
	SeasonNumber     int           `xml:"SeasonNumber"`
	Writer           PipeList      `xml:"Writer"`
	AbsoluteNumber   NullInt       `xml:"absolute_number"`
	BannerFilename   string        `xml:"filename"`
	LastUpdated      UnixEpochTime `xml:"lastupdated"` // Unix Timestamp
	SeasonID         int           `xml:"seasonid"`
	SeriesID         int           `xml:"seriesid"`
	ThumbAdded       DateTimeStamp `xml:"thumb_added"` //Date/Time
	ThumbHeight      NullInt       `xml:"thumb_height"`
	ThumbWidth       NullInt       `xml:"thumb_width"`
	// Deprecated
	//DvdChapter            int   `xml:"DVD_chapter"`
	//DvdDiscID             string   `xml:"DVD_discid"`
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

// Series represents TV show on TheTVDB.
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

// NewClient returns a new TVDB API instance.:
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

// Lanauges gets a list of lanauges currently supported on TVDB.
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

// SeriesByID gets a single series' details from the TVDB series id.
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

// UserFavs gets a list of a TVDB's user favorite series.   Returns the series
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
