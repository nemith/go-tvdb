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
	Seasons       map[int][]*Episode
	seriesShared
}

type SeriesFull struct {
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

// data is the response back from the server
type Data struct {
	Series   []*Series  `xml:"Series,omitempty"`
	Episodes []*Episode `xml:"Episode,omitempty"`
}

type TVDB struct {
	APIKey      string
	DefaultLang string
}

func NewTVDB(apiKey string) *TVDB {
	return &TVDB{
		APIKey:      apiKey,
		DefaultLang: "en",
	}
}

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

func (t *TVDB) baseURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   "thetvdb.com",
	}
}

func (t *TVDB) apiURL(path string, query url.Values) *url.URL {
	url := t.baseURL()
	url.Path = fmt.Sprintf("api/%s", path)
	url.RawQuery = query.Encode()
	return url
}

func (t *TVDB) staticAPIURL(path string) *url.URL {
	url := t.baseURL()
	url.Path = fmt.Sprintf("api/%s/%s", t.APIKey, path)
	return url
}

// GetSeries queries for a series by the series name. Returns a list of matches
// See http://thetvdb.com/wiki/index.php?title=API:GetSeries for more information
func (t *TVDB) GetSeries(name string) ([]SeriesSummary, error) {
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
func (t *TVDB) GetSeriesByID(id int) (*Series, error) {
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
func (t *TVDB) GetSeriesByRemoteID(service RemoteService, id string) (*Series, error) {
	query := url.Values{}
	query.Set(string(service), id)
	u := t.apiURL("GetSeriesByRemoteID.php", query)
	data := &Data{}
	if err := getResponse(u.String(), data); err != nil {
		return nil, err
	}

	if len(data.Series) != 1 {
		return nil, fmt.Errorf("Got too many series (expected: 1, got: %d)", len(data.Series))
	}

	return data.Series[0], nil
}

// GetSeriesFull grabs the static Full Series Record for the series by the
// series id.
// See: http://thetvdb.com/wiki/index.php?title=API:Full_Series_Record
func (t *TVDB) GetSeriesFull(seriesID int) (*Series, error) {
	u := t.staticAPIURL(fmt.Sprintf("series/%d/all/en.xml", seriesID))
	data := &Data{}
	if err := getResponse(u.String(), data); err != nil {
		return nil, err
	}

	if len(data.Series) != 1 {
		return nil, fmt.Errorf("Got too many series (expected: 1, got: %d)", len(data.Series))
	}

	series := data.Series[0]
	if series.Seasons == nil {
		series.Seasons = make(map[int][]*Episode, len(data.Episodes))
	}

	for _, episode := range data.Episodes {
		series.Seasons[episode.SeasonNumber] = append(series.Seasons[episode.SeasonNumber], episode)
	}
	return series, nil
}

var reSearchSeries = regexp.MustCompile(`(?P<before><a href="/\?tab=series&amp;id=)(?P<seriesId>\d+)(?P<after>\&amp;lid=\d*">)`)

// SearchSeries searches for TV series by name, using the user based search
// found on TVDB's homepage.
func (t *TVDB) SearchSeries(name string, maxResults int) ([]Series, error) {
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

	if len(results) < maxResults {
		maxResults = len(results)
	}
	seriesList := make([]Series, maxResults)

	for _, result := range results {
		seriesID := int64(0)
		var series *Series
		seriesID, err = strconv.ParseInt(string(result[2]), 10, 64)
		if err != nil {
			continue
		}

		series, err = t.GetSeriesByID(int(seriesID))
		if err != nil {
			// Some series can't be found, so we will ignore these.
			if _, ok := err.(*xml.SyntaxError); ok {
				err = nil
				continue
			} else {
				return seriesList, err
			}
		}

		seriesList = append(seriesList, *series)

		if len(seriesList) == maxResults {
			break
		}
	}
	return seriesList, nil
}

// userFav is the internal function for UserFav, UserFavAdd, and UserFavRemove
// since they all use the same API.
func (t *TVDB) userFav(accountID, actionType string, seriesID int) ([]int, error) {
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
func (t *TVDB) UserFav(accountID string) ([]int, error) {
	return t.userFav(accountID, "", 0)
}

// UserFavAdd will add a series by series id to a users account.  See UserFav
// for information on account id. Returns the modified list
func (t *TVDB) UserFavAdd(accountID string, seriesID int) ([]int, error) {
	return t.userFav(accountID, "add", seriesID)
}

// UserFavRemove will delete a series by series id to a users account.  See
// UserFav for information on account id. Returns the modified list
func (t *TVDB) UserFavRemove(accountID string, seriesID int) ([]int, error) {
	return t.userFav(accountID, "remove", seriesID)
}

type ratingResult struct {
	SerRatings []*Rating `xml:"Series"`
	EpRatings  []*Rating `xml:"Episode"`
}

func (t *TVDB) getRatingsForUser(accountID string, seriesID int) (*ratingResult, error) {
	query := url.Values{}

	query.Set("apikey", t.APIKey) //Love the consistency of this API
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
func (t *TVDB) GetRatingsForUser(accountID string) ([]*Rating, error) {
	result, err := t.getRatingsForUser(accountID, 0)
	if err != nil {
		return nil, err
	}

	return result.SerRatings, nil
}

// GetRatingsForUserSeries will return the user and community ratings for a
// series as well as all the episodes.  Returns the Series ratings first and
// then a slice of episode ratings.
func (t *TVDB) GetRaitingsForUserSeries(accountID string, seriesID int) (*Rating, []*Rating, error) {
	result, err := t.getRatingsForUser(accountID, seriesID)
	if err != nil {
		return nil, nil, err
	}

	return result.SerRatings[0], result.EpRatings, nil
}

// setUserRating is a commond function for both SetUserRatingSeries and
// SetUserRatingEpisode since they utilize the same API.
func (t *TVDB) setUserRating(accountID, itemType string, itemID, rating int) error {
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
func (t *TVDB) SetUserRatingSeries(accountID string, seriesID, rating int) error {
	return t.setUserRating(accountID, "series", seriesID, rating)
}

// UserRatingEp will update the user ratiing for the episode by episode id.
func (t *TVDB) SetUserRatingEp(accountID string, epID, rating int) error {
	return t.setUserRating(accountID, "episode", epID, rating)
}

// UserLang will return the prefered language for a user with a given account
// id.
func (t *TVDB) UserLang(accountID string) (*Language, error) {
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
