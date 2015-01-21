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

// UnmarshalXML unmarshals an XML element with string value into a pip-separated list of strings.
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
	ID                    uint64   `xml:"id"`
	CombinedEpisodeNumber string   `xml:"Combined_episodenumber"`
	CombinedSeason        uint64   `xml:"Combined_season"`
	DvdChapter            string   `xml:"DVD_chapter"`
	DvdDiscID             string   `xml:"DVD_discid"`
	DvdEpisodeNumber      string   `xml:"DVD_episodenumber"`
	DvdSeason             string   `xml:"DVD_season"`
	Director              PipeList `xml:"Director"`
	EpImgFlag             string   `xml:"EpImgFlag"`
	EpisodeName           string   `xml:"EpisodeName"`
	EpisodeNumber         uint64   `xml:"EpisodeNumber"`
	FirstAired            string   `xml:"FirstAired"`
	GuestStars            string   `xml:"GuestStars"`
	ImdbID                string   `xml:"IMDB_ID"`
	Language              string   `xml:"Language"`
	Overview              string   `xml:"Overview"`
	ProductionCode        string   `xml:"ProductionCode"`
	Rating                string   `xml:"Rating"`
	RatingCount           string   `xml:"RatingCount"`
	SeasonNumber          uint64   `xml:"SeasonNumber"`
	Writer                PipeList `xml:"Writer"`
	AbsoluteNumber        string   `xml:"absolute_number"`
	Filename              string   `xml:"filename"`
	LastUpdated           string   `xml:"lastupdated"`
	SeasonID              uint64   `xml:"seasonid"`
	SeriesID              uint64   `xml:"seriesid"`
	ThumbAdded            string   `xml:"thumb_added"`
	ThumbHeight           string   `xml:"thumb_height"`
	ThumbWidth            string   `xml:"thumb_width"`
}

// Series represents TV show on TheTVDB.
type Series struct {
	ID            uint64   `xml:"id"`
	Actors        PipeList `xml:"Actors"`
	AirsDayOfWeek string   `xml:"Airs_DayOfWeek"`
	AirsTime      string   `xml:"Airs_Time"`
	ContentRating string   `xml:"ContentRating"`
	FirstAired    string   `xml:"FirstAired"`
	Genre         PipeList `xml:"Genre"`
	ImdbID        string   `xml:"IMDB_ID"`
	Language      string   `xml:"Language"`
	Network       string   `xml:"Network"`
	NetworkID     string   `xml:"NetworkID"`
	Overview      string   `xml:"Overview"`
	Rating        string   `xml:"Rating"`
	RatingCount   string   `xml:"RatingCount"`
	Runtime       string   `xml:"Runtime"`
	SeriesID      string   `xml:"SeriesID"`
	SeriesName    string   `xml:"SeriesName"`
	Status        string   `xml:"Status"`
	Added         string   `xml:"added"`
	AddedBy       string   `xml:"addedBy"`
	Banner        string   `xml:"banner"`
	Fanart        string   `xml:"fanart"`
	LastUpdated   string   `xml:"lastupdated"`
	Poster        string   `xml:"poster"`
	Zap2ItID      string   `xml:"zap2it_id"`
	Seasons       map[uint64][]*Episode
}

type Language struct {
	ID   int    `xml:"id"`
	Abbr string `xml:"abbreviation"`
	Name string `xml:"name"`
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
func (t *TVDB) GetSeries(name string) ([]*Series, error) {
	u := t.apiURL("GetSeries.php", url.Values{
		"seriesname": []string{name},
	})
	data := &Data{}
	if err := getResponse(u.String(), data); err != nil {
		return nil, err
	}
	return data.Series, nil
}

// GetSeriesByID grabs the static Base Series Record file by the TVDB series id.
// See http://thetvdb.com/wiki/index.php?title=API:Base_Series_Record
func (t *TVDB) GetSeriesByID(id uint64) (*Series, error) {
	u := t.staticAPIURL(fmt.Sprintf("series/%d/en.xml", id))
	data := &Data{}
	if err := getResponse(u.String(), data); err != nil {
		return nil, err
	}

	if len(data.Series) != 1 {
		return nil, fmt.Errorf("Got too many series (expected: 1, got: %d)", len(data.Series))
	}

	return data.Series[0], nil
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
func (t *TVDB) GetSeriesFull(seriesID uint64) (*Series, error) {
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
		series.Seasons = make(map[uint64][]*Episode, len(data.Episodes))
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
		seriesID := uint64(0)
		var series *Series
		seriesID, err = strconv.ParseUint(string(result[2]), 10, 64)
		if err != nil {
			continue
		}

		series, err = t.GetSeriesByID(seriesID)
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

// userRating is a commond function for both UserRatingSeries and
// UserRatingEpisode since they utilize the same API.
func (t *TVDB) userRating(accountID, itemType string, itemID, rating int) error {
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
	if err := getResponse(u.String(), &result); err != nil {
		return err
	}
	return nil
}

// UserRatingSeries will update the user rating for the series bu the series id.
func (t *TVDB) UserRatingSeries(accountID string, seriesID, rating int) error {
	return t.userRating(accountID, "series", seriesID, rating)
}

// UserRatingEp will update the user ratiing for the episode by episode id.
func (t *TVDB) UserRatingEp(accountID string, epID, rating int) error {
	return t.userRating(accountID, "episode", epID, rating)
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
