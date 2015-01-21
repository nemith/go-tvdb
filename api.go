package tvdb

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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

// Episode represents a TV show episode on TheTVDB.
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

// Langage used for TVDB content
type Language struct {
	ID   int    `xml:"id"`
	Abbr string `xml:"abbreviation"`
	Name string `xml:"name"`
}

// Rating of a show or episode for both user rating as well as community rating
type Rating struct {
	ID              int `xml:"id"`
	UserRating      int
	CommunityRating float32
}

// Hack to combine xml feilds id and seriesid into a single field so we can use it
// for both series and episodes
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

type RemoteService string

const (
	IMDB   = RemoteService("imdbid")
	Zap2it = RemoteService("zap2it")
)

// API is the low level API for accessing thetvdb.com "api".  Most
// functions of API should mimic thier counterparts found in the
// public API with some fetching and parsing thrown in
type API struct {
	Key         string
	DefaultLang string
}

// NewAPI creates a new API instance with an api key.  Language defaults
// to English.
func NewAPI(key string) *API {
	return &API{
		Key:         key,
		DefaultLang: "en",
	}
}

// getResponse is a helper function to fetch and parse xml data from thetvdb.com
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

// baseURL is used to generate the basic URL for thetvdb.com
func (t *API) baseURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   "thetvdb.com",
	}
}

// apiURL buolds on baseURL and provides a quick utility for generating a url
// to the dyanamic calls to the TVDB API (i.e the PHP scripts)
func (t *API) apiURL(path string, query url.Values) *url.URL {
	url := t.baseURL()
	url.Path = fmt.Sprintf("api/%s", path)
	url.RawQuery = query.Encode()
	return url
}

// staticURL builds on base use and provides a quick utility for generating a
// url to static parts of the TVDB API (Static zip and xml files)
func (t *API) staticAPIURL(path string) *url.URL {
	url := t.baseURL()
	url.Path = fmt.Sprintf("api/%s/%s", t.Key, path)
	return url
}

// GetSeries queries for a series by the series name. Returns a list of matches
// See http://thetvdb.com/wiki/index.php?title=API:GetSeries for more information
func (t *API) GetSeries(name string) ([]SeriesSummary, error) {
	u := t.apiURL("GetSeries.php", url.Values{
		"seriesname": []string{name},
	})
	response := struct {
		XMLName xml.Name `xml:"Data"`
		Series  []SeriesSummary
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}
	return response.Series, nil
}

// GetSeriesByID grabs the static Base Series Record file by the TVDB series id.
// See http://thetvdb.com/wiki/index.php?title=API:Base_Series_Record
func (t *API) GetSeriesByID(id int) (*Series, error) {
	u := t.staticAPIURL(fmt.Sprintf("series/%d/en.xml", id))
	response := struct {
		XMLName xml.Name `xml:"Data"`
		Series  Series
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}

	return &response.Series, nil
}

// GetSeriesByRemoteID queries the tvdb database for a series based on a remote
// id.  The RemoteID is the identifier used by a remote system like IMDB or
// Zap2it.
// See: http://thetvdb.com/wiki/index.php?title=API:GetSeriesByRemoteID
func (t *API) GetSeriesByRemoteID(service RemoteService, id string) (*Series, error) {
	query := url.Values{}
	query.Set(string(service), id)
	u := t.apiURL("GetSeriesByRemoteID.php", query)
	response := struct {
		XMLName xml.Name `xml:"Data"`
		Series  Series
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}

	return &response.Series, nil
}

// GetSeriesFull grabs the static Full Series Record for the series by the
// series id.
// See: http://thetvdb.com/wiki/index.php?title=API:Full_Series_Record
func (t *API) GetSeriesEp(seriesID int) (*Series, error) {
	u := t.staticAPIURL(fmt.Sprintf("series/%d/all/en.xml", seriesID))
	response := struct {
		XMLName  xml.Name `xml:"Data"`
		Series   Series
		Episodes []Episode `xml:"Episode"`
	}{}
	if err := getResponse(u.String(), &response); err != nil {
		return nil, err
	}

	//if series.Seasons == nil {
	//	series.Seasons = make(map[int][]*Episode, len(data.Episodes))
	//}

	//for _, episode := range data.Episodes {
	//	series.Seasons[episode.SeasonNumber] = append(series.Seasons[episode.SeasonNumber], episode)
	//}
	return &response.Series, nil
}

var reSearchSeries = regexp.MustCompile(`<a href="/\?tab=series&amp;id=(\d+)\&amp;lid=\d*">`)

// SearchSeries searches for TV series by name, using the user based search
// found on TVDB's homepage.
func (t *API) SearchSeries(name string) ([]int, error) {
	u := t.baseURL()
	query := url.Values{
		"string":         []string{name},
		"searchseriesid": []string{""},
		"tab":            []string{"listseries"},
		"function":       []string{"Search"},
	}
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	results := reSearchSeries.FindAllStringSubmatch(buf.String(), -1)
	seriesList := make([]int, len(results))

	for _, result := range results {
		seriesID, err := strconv.ParseInt(string(result[1]), 10, 64)
		if err != nil {
			continue
		}
		seriesList = append(seriesList, int(seriesID))
	}
	return seriesList, nil
}

// userFav is the internal function for UserFav, UserFavAdd, and UserFavRemove
// since they all use the same API.
func (t *API) userFav(accountID, actionType string, seriesID int) ([]int, error) {
	query := url.Values{}
	query.Set("accountid", accountID)

	if actionType != "" {
		query.Set("type", actionType)
		query.Set("seriesid", strconv.FormatInt(int64(seriesID), 10))
	}

	u := t.apiURL("User_Favorites.php", query)

	data := &struct {
		XMLName xml.Name `xml:"Favorites"`
		Series  []int
	}{}

	if err := getResponse(u.String(), data); err != nil {
		return nil, err
	}
	return data.Series, nil
}

// UserFav queries TVDB's database for favorites for a given accound id. Please
// note this is the accountID and not the username of the account.  Users can
// find thier account id from thier account page
// (http://thetvdb.com/?tab=userinfo).
// Returns a slice of series ids
func (t *API) UserFav(accountID string) ([]int, error) {
	return t.userFav(accountID, "", 0)
}

// UserFavAdd will add a series by series id to a users account.  See UserFav
// for information on account id. Returns the modified list
func (t *API) UserFavAdd(accountID string, seriesID int) ([]int, error) {
	return t.userFav(accountID, "add", seriesID)
}

// UserFavRemove will delete a series by series id to a users account.  See
// UserFav for information on account id. Returns the modified list
func (t *API) UserFavRemove(accountID string, seriesID int) ([]int, error) {
	return t.userFav(accountID, "remove", seriesID)
}

type ratingResult struct {
	SerRatings []*Rating `xml:"Series"`
	EpRatings  []*Rating `xml:"Episode"`
}

func (t *API) getRatingsForUser(accountID string, seriesID int) (*ratingResult, error) {
	query := url.Values{}

	query.Set("apikey", t.Key) //Love the consistency of this API
	query.Set("accountid", accountID)
	if seriesID != 0 {
		query.Set("seriesid", strconv.FormatInt(int64(seriesID), 10))
	}
	u := t.apiURL("GetRatingsForUser.php", query)
	result := &ratingResult{}
	if err := getResponse(u.String(), result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetRatingsForUser will get all series raiting for user as well as the
// community ratings
func (t *API) GetRatingsForUser(accountID string) ([]*Rating, error) {
	result, err := t.getRatingsForUser(accountID, 0)
	if err != nil {
		return nil, err
	}

	return result.SerRatings, nil
}

// GetRatingsForUserSeries will return the user and community ratings for a
// series as well as all the episodes.  Returns the Series ratings first and
// then a slice of episode ratings.
func (t *API) GetRaitingsForUserSeries(accountID string, seriesID int) (*Rating, []*Rating, error) {
	result, err := t.getRatingsForUser(accountID, seriesID)
	if err != nil {
		return nil, nil, err
	}

	return result.SerRatings[0], result.EpRatings, nil
}

// setUserRating is a commond function for both SetUserRatingSeries and
// SetUserRatingEpisode since they utilize the same API.
func (t *API) setUserRating(accountID, itemType string, itemID, rating int) error {
	if rating < 0 || rating > 10 {
		return fmt.Errorf("Rating must be between 0 and 10 inclusive")
	}

	query := url.Values{}
	query.Set("accountid", accountID)
	query.Set("itemtype", itemType)
	query.Set("itemid", strconv.FormatInt(int64(itemID), 10))
	query.Set("rating", strconv.FormatInt(int64(rating), 10))
	u := t.apiURL("User_Rating.php", query)

	// Result is the site rating for some reason.  The API on this site is wack
	result := &struct{}{}
	if err := getResponse(u.String(), result); err != nil {
		return err
	}
	return nil
}

// UserRatingSeries will update the user rating for the series bu the series id.
func (t *API) SetUserRatingSeries(accountID string, seriesID, rating int) error {
	return t.setUserRating(accountID, "series", seriesID, rating)
}

// UserRatingEp will update the user ratiing for the episode by episode id.
func (t *API) SetUserRatingEp(accountID string, epID, rating int) error {
	return t.setUserRating(accountID, "episode", epID, rating)
}

// UserLang will return the prefered language for a user with a given account
// id.
func (t *API) UserLang(accountID string) (*Language, error) {
	u := t.apiURL("User_PreferredLanguage.php", url.Values{
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
