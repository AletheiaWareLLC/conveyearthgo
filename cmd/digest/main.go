package main

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/content/markdown"
	"aletheiaware.com/conveyearthgo/content/plaintext"
	"aletheiaware.com/netgo"
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/transform"
	"github.com/bmaupin/go-epub"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goki/freetype/truetype"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	DIGEST_LIMIT = 6
	NEW_LINE     = "\n"
)

var (
	year    = flag.Int("year", 2020, "Year")
	month   = flag.Int("month", 1, "Month")
	textBG  = flag.Int("textbg", 0x8f, "Text Background")
	colors  = flag.String("colors", "#00bfff,#87cefa", "Color Scheme (Primary,Secondary)")
	fonts   = flag.String("fonts", "fonts", "Fonts directory")
	uploads = flag.String("uploads", "uploads", "Uploads directory")
	edits   = flag.String("edits", "edits", "Edits directory")
	emojis  = flag.String("emojis", "emojis", "Emojis directory")
	reward  = flag.Float64("reward", 12500., "Reward (cents)")
	//    625 - 2021
	//    625 - 2022
	//   1250 - 2023
	//   2500 - 2024
	//   5000 - 2025
	//  10000 - 2026
	//  20000 - 2027
	//  40000 - 2028
	//  80000 - 2029
	// 160000 - 2030

	//go:embed assets
	embeddedFS embed.FS

	titleFont   = "NotoSerif-ExtraBold.ttf"
	editionFont = "NotoSerif-Bold.ttf"
	topicFont   = "NotoSerif-Regular.ttf"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Missing output directory")
	}
	output := args[0]

	host, ok := os.LookupEnv("HOST")
	if !ok {
		log.Fatal(errors.New("Missing HOST environment variable"))
	}
	scheme := "http"
	if netgo.IsSecure() {
		scheme = "https"
	}
	title := "Convey"
	description := "Quality Public Dialogue"
	author := "Stuart Scott"
	width := 1600
	height := 2560

	start := time.Date(*year, time.Month(*month), 1, 0, 0, 0, 0, time.UTC)
	log.Println("Start:", start, start.Unix())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	log.Println("End:", end, end.Unix())

	parts := strings.Split(*colors, ",")
	primary, err := parseColor(parts[0])
	if err != nil {
		log.Fatal(err)
	}
	secondary, err := parseColor(parts[1])
	if err != nil {
		log.Fatal(err)
	}

	e := epub.NewEpub(title)
	e.SetAuthor(author)
	e.SetDescription(description)
	if _, err := e.AddFont(path.Join(*fonts, "NotoSerif-Regular.ttf"), ""); err != nil {
		log.Fatal(err)
	}

	cover := &Cover{
		Width:          width,
		Height:         height,
		TextBackground: uint8(*textBG),
		Emojis:         *emojis,
		Title:          strings.ToUpper(title),
		TitleSize:      256,
		TitleColor:     primary,
		Edition: [2]string{
			start.Format("2006"),
			strings.ToUpper(start.Format("January")),
		},
		EditionSize:  36,
		EditionColor: secondary,
		TopicSize:    60,
		TopicColor:   primary,
	}

	coverHasBackgroundImage := false

	// Parse Templates
	templateFS, err := fs.Sub(embeddedFS, "assets")
	if err != nil {
		log.Fatal(err)
	}
	templates, err := template.ParseFS(templateFS, "*.go.*")
	if err != nil {
		log.Fatal(err)
	}

	// Copy CSS from assets to temp file
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	bodyCss, err := os.Create(path.Join(tempDir, fmt.Sprintf(`body-%s-%s.css`, start.Format("2006"), start.Format("01"))))
	if err != nil {
		log.Fatal(err)
	}
	defer bodyCss.Close()

	if err := templates.ExecuteTemplate(bodyCss, "body.go.css", struct {
		Primary   string
		Secondary string
	}{
		Primary:   parts[0],
		Secondary: parts[1],
	}); err != nil {
		log.Fatal(err)
	}

	internalBodyCss, err := e.AddCSS(bodyCss.Name(), "")
	if err != nil {
		log.Fatal(err)
	}

	db, err := openDatabase()
	if err != nil {
		log.Fatal(err)
	}

	yields := make(map[string]int64)

	// Lookup Conversations
	i := 0
	if err := queryConversations(db, start, end, func(conversation, user int64, username string, topic string, created time.Time, cost, yield int64) error {
		log.Println(conversation, user, username, topic, created, cost, yield)

		cover.Topics[i] = topic

		c := &Conversation{
			ConversationId: conversation,
			Topic:          topic,
		}
		c.User = user
		c.Username = username
		c.Created = created
		c.Cost = cost
		c.Yield = yield
		ms := make(map[int64]*Message)

		// Lookup Messages
		if err := queryMessages(db, conversation, start, end, func(message, user int64, username string, parent int64, created time.Time, cost, yield int64) error {
			log.Println(message, user, username, parent, created, cost, yield)

			var m *Message
			if parent == 0 {
				m = &c.Message
			} else {
				m = &Message{}
				ms[message] = m
			}
			m.MessageId = message
			m.User = user
			m.Username = username
			m.Parent = parent
			m.Created = created
			m.Cost = cost
			m.Yield = yield

			// Lookup Files
			if err := queryFiles(db, message, func(file int64, hash, mime string, created time.Time) error {
				log.Println(file, hash, mime, created)
				m.Files = append(m.Files, &File{
					Id:      file,
					Hash:    hash,
					Mime:    mime,
					Created: created,
				})
				// TODO FIXME this will pick the first picture, regardless of whether it is in the main thread that ends up in the digest
				// TODO Ideally the cover image can be selected by passing the hash as a flag
				if !coverHasBackgroundImage {
					switch mime {
					case conveyearthgo.MIME_IMAGE_JPEG,
						conveyearthgo.MIME_IMAGE_JPG,
						conveyearthgo.MIME_IMAGE_PNG:
						src, err := os.Open(uploadPath(hash))
						if err != nil {
							return err
						}
						img, _, err := image.Decode(src)
						if err != nil {
							return err
						}
						cover.Background = img
						coverHasBackgroundImage = true
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Fatal(err)
		}

		// Set Replies
		for _, m := range ms {
			if m.Parent == c.MessageId {
				c.Replies = append(c.Replies, m)
			} else {
				p := ms[m.Parent]
				p.Replies = append(p.Replies, m)
			}
		}
		// Sort Replies
		sort.Slice(c.Replies, func(i, j int) bool {
			return c.Replies[i].Yield > c.Replies[j].Yield
		})
		for _, m := range ms {
			sort.Slice(m.Replies, func(i, j int) bool {
				return m.Replies[i].Yield > m.Replies[j].Yield
			})
		}

		for m := &c.Message; ; m = m.Replies[0] {
			yields[m.Username] = yields[m.Username] + m.Yield
			if len(m.Replies) == 0 {
				break
			}
		}

		body := fmt.Sprintf(`<h1 class="title"><a href="%s://%s/conversation?id=%d">%s</a></h1>%s`, scheme, host, c.ConversationId, topic, NEW_LINE)
		s, err := messageToHTML(e, &c.Message)
		if err != nil {
			return err
		}
		body += s
		body += fmt.Sprintf(`<a class="more" href="%s://%s/conversation?id=%d">More</a>%s`, scheme, host, c.ConversationId, NEW_LINE)
		e.AddSection(body, topic, "", internalBodyCss)

		i++
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if netgo.IsLive() {
		var authors []string
		for author := range yields {
			authors = append(authors, author)
		}
		sort.Slice(authors, func(i, j int) bool {
			return yields[authors[i]] > yields[authors[j]]
		})
		var sum int64
		for _, author := range authors {
			if !isMemberOrPartner(author) {
				sum += yields[author]
			}
		}
		rewardBody := `<h1 class="title">Rewards</h1>` + NEW_LINE
		if sum == 0 {
			rewardBody += fmt.Sprintf(`<p>The $%.2f prize fund was not claimed this month and will instead be rolled over into next month's edition.</p>%s`, *reward/100., NEW_LINE)
		} else {
			rewardTable := `<table style="text-align: center;width: 100%;">` + NEW_LINE
			rewardTable += `<tr><th>Author</th><th>Yield</th><th>Reward</th></tr>` + NEW_LINE
			var (
				total              float64
				hasMemberOrPartner bool
			)
			for _, author := range authors {
				yield := yields[author]
				if isMemberOrPartner(author) {
					rewardTable += fmt.Sprintf(`<tr><td>%s</td><td>%d</td><td>$0.00 *</td></tr>%s`, author, yield, NEW_LINE)
					hasMemberOrPartner = true
				} else {
					amount := (float64(yield) / float64(sum)) * *reward
					// Round up to nearest cent
					amount = math.Ceil(amount)
					// Convert cents to dollars
					amount /= 100.
					total += amount
					rewardTable += fmt.Sprintf(`<tr><td>%s</td><td>%d</td><td>$%.2f</td></tr>%s`, author, yield, amount, NEW_LINE)
				}
			}
			rewardTable += fmt.Sprintf(`<tr><th>Total</th><td>%d</td><td>$%.2f</td></tr>%s`, sum, total, NEW_LINE)
			rewardTable += `</table>` + NEW_LINE

			rewardBody += fmt.Sprintf(`<p>The $%.2f prize fund was shared proportionally between the authors featured in this edition based on the yield of their content, as shown below.</p>%s`, total, NEW_LINE)
			rewardBody += rewardTable
			if hasMemberOrPartner {
				rewardBody += fmt.Sprintf(`<p>* Aletheia Ware members and partners are ineligible for rewards.</p>%s`, NEW_LINE)
			}
		}
		e.AddSection(rewardBody, "Rewards", "", internalBodyCss)
	}

	aboutBody := `<h1 class="title">About</h1>` + NEW_LINE
	aboutBody += fmt.Sprintf(`<p><a href="%s://%s">Convey</a> is a Communication Platform that Incentivizes Quality Content, Collaboration, and Discussion.</p>%s`, scheme, host, NEW_LINE)
	aboutBody += fmt.Sprintf(`<p>This edition of the Convey Digest contains the highest yielding contributions from %s.</p>%s`, start.Format("January 2006"), NEW_LINE)
	aboutBody += `<p>Convey is made available by <a href="https://aletheiaware.com">Aletheia Ware</a> under the <a href="https://aletheiaware.com/terms-of-service.html">Terms of Service</a> and <a href="https://aletheiaware.com/privacy-policy.html">Privacy Policy</a>.</p>` + NEW_LINE
	aboutBody += `<p>Convey is an open-source project released under the <a href="http://www.apache.org/licenses/LICENSE-2.0">Apache 2.0 License</a> and hosted on <a href="https://github.com/AletheiaWareLLC">Github</a>.</p>` + NEW_LINE

	e.AddSection(aboutBody, "About", "", internalBodyCss)

	var logoSvg bytes.Buffer
	if err := templates.ExecuteTemplate(&logoSvg, "logo.go.svg", struct {
		Color string
	}{
		Color: parts[1],
	}); err != nil {
		log.Fatal(err)
	}
	icon, err := oksvg.ReadIconStream(&logoSvg)
	if err != nil {
		log.Fatal(err)
	}
	logo := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	scanner := rasterx.NewScannerGV(int(icon.ViewBox.W), int(icon.ViewBox.H), logo, logo.Bounds())
	raster := rasterx.NewDasher(100, 100, scanner)
	icon.SetTarget(0, 0, 100, 100)
	icon.Draw(raster, 1)
	cover.Logo = logo

	coverTitleFont, err := loadFont(path.Join(*fonts, titleFont))
	if err != nil {
		log.Fatal(err)
	}
	cover.TitleFont = coverTitleFont

	coverEditionFont, err := loadFont(path.Join(*fonts, editionFont))
	if err != nil {
		log.Fatal(err)
	}
	cover.EditionFont = coverEditionFont

	coverTopicFont, err := loadFont(path.Join(*fonts, topicFont))
	if err != nil {
		log.Fatal(err)
	}
	cover.TopicFont = coverTopicFont

	if !coverHasBackgroundImage {
		log.Fatal("Cover Has No Background Image")
	}

	coverImage := cover.Image()
	coverPng, err := os.Create(path.Join(output, fmt.Sprintf(`Convey-Digest-%s-%s.png`, start.Format("2006"), start.Format("01"))))
	if err != nil {
		log.Fatal(err)
	}
	defer coverPng.Close()

	if err := png.Encode(coverPng, coverImage); err != nil {
		log.Fatal(err)
	}

	internalCoverImage, err := e.AddImage(coverPng.Name(), "")
	if err != nil {
		log.Fatal(err)
	}

	// Write out the cover in multiple smaller sizes
	ratio := float64(height) / float64(width)
	for _, w := range []int{240, 480, 1024} {
		out, err := os.Create(path.Join(output, fmt.Sprintf(`Convey-Digest-%s-%s-%dw.png`, start.Format("2006"), start.Format("01"), w)))
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		if err := png.Encode(out, transform.Resize(coverImage, w, int(float64(w)*ratio), transform.Gaussian)); err != nil {
			log.Fatal(err)
		}
	}

	// Copy CSS from assets to temp file
	coverCss, err := os.Create(path.Join(tempDir, fmt.Sprintf(`cover-%s-%s.css`, start.Format("2006"), start.Format("01"))))
	if err != nil {
		log.Fatal(err)
	}
	defer coverCss.Close()

	coverCssFile, err := embeddedFS.Open(path.Join("assets", "cover.css"))
	if err != nil {
		log.Fatal(err)
	}
	defer coverCssFile.Close()

	if _, err := io.Copy(coverCss, coverCssFile); err != nil {
		log.Fatal(err)
	}

	internalCoverCss, err := e.AddCSS(coverCss.Name(), "")
	if err != nil {
		log.Fatal(err)
	}
	e.SetCover(internalCoverImage, internalCoverCss)

	if err := e.Write(path.Join(output, fmt.Sprintf(`Convey-Digest-%s-%s.epub`, start.Format("2006"), start.Format("01")))); err != nil {
		log.Fatal(err)
	}
}

type Conversation struct {
	Message
	ConversationId int64
	Topic          string
}

type Message struct {
	MessageId int64
	User      int64
	Username  string
	Created   time.Time
	Cost      int64
	Yield     int64
	Parent    int64
	Files     []*File
	Replies   []*Message
}

type File struct {
	Id      int64
	Hash    string
	Mime    string
	Created time.Time
}

func messageToHTML(e *epub.Epub, m *Message) (string, error) {
	ss := fmt.Sprintf(`<p class="author">%s</p>%s`, m.Username, NEW_LINE)
	for _, f := range m.Files {
		s, err := fileToHTML(e, f)
		if err != nil {
			return "", err
		}
		ss += s
	}
	if len(m.Replies) > 0 {
		// Only the reply with the highest yield is included
		s, err := messageToHTML(e, m.Replies[0])
		if err != nil {
			return "", err
		}
		ss += s
	}
	return ss, nil
}

func fileToHTML(e *epub.Epub, f *File) (string, error) {
	switch f.Mime {
	case conveyearthgo.MIME_IMAGE_JPEG,
		conveyearthgo.MIME_IMAGE_JPG,
		conveyearthgo.MIME_IMAGE_GIF,
		conveyearthgo.MIME_IMAGE_PNG,
		conveyearthgo.MIME_IMAGE_SVG,
		conveyearthgo.MIME_IMAGE_WEBP:
		ext, err := imageMimeToExt(f.Mime)
		if err != nil {
			return "", err
		}
		imageName, err := e.AddImage(uploadPath(f.Hash), f.Hash+ext)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`<img class="ucc" src="%s"/>%s`, imageName, NEW_LINE), nil
	case conveyearthgo.MIME_TEXT_PLAIN:
		file, err := os.Open(uploadPath(f.Hash))
		if err != nil {
			return "", err
		}
		html, err := plaintext.ToHTML(file)
		if err != nil {
			return "", err
		}
		return string(html) + NEW_LINE, nil
	case conveyearthgo.MIME_TEXT_MARKDOWN:
		file, err := os.Open(uploadPath(f.Hash))
		if err != nil {
			return "", err
		}
		html, err := markdown.ToHTML(file)
		if err != nil {
			return "", err
		}
		return string(html), nil
	case conveyearthgo.MIME_VIDEO_MP4,
		conveyearthgo.MIME_VIDEO_OGG,
		conveyearthgo.MIME_VIDEO_WEBM:
		ext, err := videoMimeToExt(f.Mime)
		if err != nil {
			return "", err
		}
		videoName, err := e.AddVideo(uploadPath(f.Hash), f.Hash+ext)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`<video class="ucc" controls="controls"><source src="%s" type="%s"/></video>%s`, videoName, f.Mime, NEW_LINE), nil
	}
	return "", conveyearthgo.ErrMimeUnrecognized
}

func imageMimeToExt(mime string) (string, error) {
	switch mime {
	case conveyearthgo.MIME_IMAGE_JPEG,
		conveyearthgo.MIME_IMAGE_JPG:
		return ".jpg", nil
	case conveyearthgo.MIME_IMAGE_GIF:
		return ".gif", nil
	case conveyearthgo.MIME_IMAGE_PNG:
		return ".png", nil
	case conveyearthgo.MIME_IMAGE_SVG:
		return ".svg", nil
	case conveyearthgo.MIME_IMAGE_WEBP:
		return ".webp", nil
	}
	return "", conveyearthgo.ErrMimeUnrecognized
}

func videoMimeToExt(mime string) (string, error) {
	switch mime {
	case conveyearthgo.MIME_VIDEO_MP4:
		return ".mp4", nil
	case conveyearthgo.MIME_VIDEO_OGG:
		return ".ogg", nil
	case conveyearthgo.MIME_VIDEO_WEBM:
		return ".webm", nil
	}
	return "", conveyearthgo.ErrMimeUnrecognized
}

func openDatabase() (*sql.DB, error) {
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dsn := fmt.Sprintf(`%s:%s@tcp(localhost:3306)/%s`, dbUser, dbPassword, dbName)
	return sql.Open("mysql", dsn)
}

func queryConversations(db *sql.DB, start, end time.Time, callback func(int64, int64, string, string, time.Time, int64, int64) error) error {
	rows, err := db.Query(`
        SELECT tbl_conversations.id, tbl_conversations.user, tbl_users.username, tbl_conversations.topic, tbl_conversations.created_unix, tbl_charges.amount, IFNULL(yields.yield,0)
        FROM tbl_conversations
        INNER JOIN tbl_users ON tbl_conversations.user=tbl_users.id
        INNER JOIN tbl_messages ON tbl_conversations.id=tbl_messages.conversation AND tbl_messages.parent IS NULL
        INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
        LEFT JOIN (
            SELECT parent, SUM(IFNULL(amount,0)) AS yield
            FROM tbl_yields
            WHERE tbl_yields.created_unix BETWEEN ? AND ?
            GROUP BY parent
        ) AS yields ON tbl_messages.id=yields.parent
        WHERE yields.yield > 0 AND tbl_conversations.created_unix BETWEEN ? AND ?
        ORDER BY yields.yield DESC
        LIMIT ?`, start.Unix(), end.Unix(), start.Unix(), end.Unix(), DIGEST_LIMIT)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id       int64
			user     int64
			username string
			topic    string
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &topic, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, user, username, topic, time.Unix(created, 0), cost, yield); err != nil {
			return err
		}
	}
	return rows.Err()
}

func queryMessages(db *sql.DB, conversation int64, start, end time.Time, callback func(int64, int64, string, int64, time.Time, int64, int64) error) error {
	rows, err := db.Query(`
        SELECT tbl_messages.id, tbl_messages.user, tbl_users.username, IFNULL(tbl_messages.parent, 0), tbl_messages.created_unix, tbl_charges.amount, IFNULL(yields.yield, 0)
        FROM tbl_messages
        INNER JOIN tbl_users ON tbl_messages.user=tbl_users.id
        INNER JOIN tbl_charges ON tbl_messages.id=tbl_charges.message
        LEFT JOIN (
            SELECT parent, SUM(amount) AS yield
            FROM tbl_yields
            WHERE tbl_yields.created_unix BETWEEN ? AND ?
            GROUP BY parent
        ) AS yields ON tbl_messages.id=yields.parent
        WHERE yields.yield > 0 AND tbl_messages.conversation=?`, start.Unix(), end.Unix(), conversation)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id       int64
			user     int64
			username string
			parent   int64
			created  int64
			cost     int64
			yield    int64
		)
		if err := rows.Scan(&id, &user, &username, &parent, &created, &cost, &yield); err != nil {
			return err
		}
		if err := callback(id, user, username, parent, time.Unix(created, 0), cost, yield); err != nil {
			return err
		}
	}
	return rows.Err()
}

func queryFiles(db *sql.DB, message int64, callback func(int64, string, string, time.Time) error) error {
	rows, err := db.Query(`
        SELECT tbl_files.id, tbl_files.hash, tbl_files.mime, tbl_files.created_unix
        FROM tbl_files
        WHERE tbl_files.message=?`, message)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id      int64
			hash    string
			mime    string
			created int64
		)
		if err := rows.Scan(&id, &hash, &mime, &created); err != nil {
			return err
		}
		if err := callback(id, hash, mime, time.Unix(created, 0)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func loadFont(name string) (*truetype.Font, error) {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return truetype.Parse(data)
}

func uploadPath(hash string) string {
	s := path.Join(*edits, hash)
	if _, err := os.Stat(s); err != nil {
		s = path.Join(*uploads, hash)
	}
	log.Println("Selecting:", s)
	return s
}

func isMemberOrPartner(author string) bool {
	switch author {
	case "stuartscott", "winksaville":
		return true
	}
	return false
}

func parseColor(s string) (color.Color, error) {
	var c color.NRGBA
	var err error
	if len(s) == 7 {
		c.A = 0xFF
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	} else {
		_, err = fmt.Sscanf(s, "#%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	}
	return c, err
}
