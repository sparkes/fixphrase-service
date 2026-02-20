package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"github.com/joho/godotenv"
)

// ---- version stamping (set via -ldflags) ----
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Config struct {
	Addr       string
	ServerName string
	RepoURL    string
	GHCRImage  string
}

//go:embed wordlist/wordlist.json
var embeddedFS embed.FS

type Service struct {
	words    []string
	wordToIndex map[string]int
}
// ---------- Algorithm types ----------
type EncodeResponse struct {
	Lat     float64  `json:"lat"`
	Lon     float64  `json:"lon"`
	Phrase  string   `json:"phrase"`
	Words   []string `json:"words"`
}

type DecodeResponse struct {
	InputWords      []string `json:"inputWords"`
	CanonicalPhrase string   `json:"canonicalPhrase"`
	Lat             float64  `json:"lat"`
	Lon             float64  `json:"lon"`
	AccuracyDegrees float64  `json:"accuracyDegrees"`
}

// ---------- HTTP types ----------
type EncodeRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type DecodeRequest struct {
	Phrase string `json:"phrase"`
}

// ---------- Service setup ----------
// loadConfig reads configuration from environment variables, with defaults.
func loadConfig() Config {
	// Load .env if present (ignore error if missing)
	_ = godotenv.Load()

	cfg := Config{
		Addr:       getEnv("ADDR", ":7080"),
		ServerName: getEnv("SERVER_NAME", "fixphrase"),
		RepoURL:    getEnv("REPO_URL", ""),
		GHCRImage:  getEnv("GHCR_IMAGE", ""),
	}
	return cfg
}

// getEnv reads an environment variable and returns it, or a fallback if not set or empty.
func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

// loadEmbeddedWordlist reads the embedded wordlist.json file and returns the list of words.
func loadEmbeddedWordlist() ([]string, error) {
	b, err := embeddedFS.ReadFile("wordlist/wordlist.json")
	if err != nil {
		return nil, fmt.Errorf("read embedded wordlist.json: %w", err)
	}
	var words []string
	if err := json.Unmarshal(b, &words); err != nil {
		return nil, fmt.Errorf("decode embedded wordlist.json: %w", err)
	}
	if len(words) == 0 {
		return nil, errors.New("wordlist empty")
	}
	return words, nil
}

// newService creates a new Service instance with the provided word list
func newService(words []string) (*Service, error) {
	// Indices used by algorithm:
	// 0..1999, 2000..5609, 5610..6609, 6610..7609 => need 7610 words.
	// ignore longer wordlists, but error on shorter ones.
	if len(words) < 7610 {
		return nil, fmt.Errorf("wordlist too short: got %d, need 7610", len(words))
	}
	m := make(map[string]int, len(words))
	for i, w := range words {
		m[strings.ToLower(w)] = i
	}
	return &Service{words: words, wordToIndex: m}, nil
}

func (s *Service) word(index int) (string, error) {
	if index < 0 || index >= len(s.words) {
		return "", fmt.Errorf("word index out of range: %d (size=%d)", index, len(s.words))
	}
	return s.words[index], nil
}

// ---------- Algorithm ----------
// encodeFixPhrase takes latitude and longitude, 
// validates them, 
// and returns the resulting phrase and words.
func encodeFixPhrase(svc *Service, latitude, longitude float64) (*EncodeResponse, error) {
	if latitude > 90 || latitude < -90 {
		return nil, fmt.Errorf("latitude out of range: %v", latitude)
	}
	if longitude > 180 || longitude < -180 {
		return nil, fmt.Errorf("longitude out of range: %v", longitude)
	}

	latInt := int(math.Round(latitude*10000.0)) + 90*10000
	lonInt := int(math.Round(longitude*10000.0)) + 180*10000

	lat := fmt.Sprintf("%07d", latInt)
	lon := fmt.Sprintf("%07d", lonInt)

	lat1dec, _ := strconv.Atoi(lat[0:4])
	lon1dec, _ := strconv.Atoi(lon[0:4])
	latlon2dec, _ := strconv.Atoi(lat[4:6] + lon[4:5])
	latlon4dec, _ := strconv.Atoi(lat[6:7] + lon[5:7])

	g0 := lat1dec
	g1 := lon1dec + 2000
	g2 := latlon2dec + 5610
	g3 := latlon4dec + 6610

	w0, err := svc.word(g0)
	if err != nil {
		return nil, fmt.Errorf("word0(%d): %w", g0, err)
	}
	w1, err := svc.word(g1)
	if err != nil {
		return nil, fmt.Errorf("word1(%d): %w", g1, err)
	}
	w2, err := svc.word(g2)
	if err != nil {
		return nil, fmt.Errorf("word2(%d): %w", g2, err)
	}
	w3, err := svc.word(g3)
	if err != nil {
		return nil, fmt.Errorf("word3(%d): %w", g3, err)
	}

	words := []string{w0, w1, w2, w3}
	return &EncodeResponse{
		Lat:     latitude,
		Lon:     longitude,
		Phrase:  strings.Join(words, " "),
		Words:   words,
	}, nil
}

// decodeFixPhrase takes a phrase, 
// and returns the decoded coordinates along with the input words and a canonical phrase.
func decodeFixPhrase(svc *Service, phrase string) (*DecodeResponse, error) {
	phrase = strings.ToLower(strings.TrimSpace(phrase))
	if phrase == "" {
		return nil, errors.New("empty phrase (need at least 2 words)")
	}
	parts := strings.Fields(phrase)
	if len(parts) < 2 {
		return nil, errors.New("not enough words (need at least 2)")
	}

	indexes := []int{-1, -1, -1, -1}
	canonical := []string{"", "", "", ""}

	for _, w := range parts {
		ix, ok := svc.wordToIndex[w]
		if !ok {
			continue
		}
		switch {
		case ix >= 0 && ix < 2000:
			indexes[0] = ix
			canonical[0] = svc.words[ix]
		case ix >= 2000 && ix < 5610:
			indexes[1] = ix - 2000
			canonical[1] = svc.words[ix]
		case ix >= 5610 && ix < 6610:
			indexes[2] = ix - 5610
			canonical[2] = svc.words[ix]
		case ix >= 6610 && ix < 7610:
			indexes[3] = ix - 6610
			canonical[3] = svc.words[ix]
		}
	}

	if indexes[0] == -1 || indexes[1] == -1 {
		return nil, errors.New("supplied words input error?  This phrase is not decodable.")
	}

	divby := 10.0
	lat := fmt.Sprintf("%04d", indexes[0])
	lon := fmt.Sprintf("%04d", indexes[1])

	var latlon2dec string
	if indexes[2] != -1 {
		divby = 100.0
		latlon2dec = fmt.Sprintf("%03d", indexes[2])
		lat += latlon2dec[0:1]
		lon += latlon2dec[2:3]
	}
	if indexes[2] != -1 && indexes[3] != -1 {
		divby = 10000.0
		latlon4dec := fmt.Sprintf("%03d", indexes[3])
		lat += latlon2dec[1:2] + latlon4dec[0:1]
		lon += latlon4dec[1:3]
	}

	latNum, _ := strconv.Atoi(lat)
	lonNum, _ := strconv.Atoi(lon)

	latitude := math.Round((((float64(latNum)/divby)-90.0)*10000.0)) / 10000.0
	longitude := math.Round((((float64(lonNum)/divby)-180.0)*10000.0)) / 10000.0

	accuracy := 0.0001
	switch divby {
	case 10.0:
		latitude += 0.05
		longitude += 0.05
		accuracy = 0.1
	case 100.0:
		latitude += 0.005
		longitude += 0.005
		accuracy = 0.01
	}

	canon := strings.TrimSpace(strings.Join(canonical, " "))
	return &DecodeResponse{
		InputWords:      parts,
		CanonicalPhrase: canon,
		Lat:             latitude,
		Lon:             longitude,
		AccuracyDegrees: accuracy,
	}, nil
}

// ---------- HTTP ----------
// writeJSON is a helper function to write a JSON response
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// readJSON is a helper function to read and decode a JSON request body
func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// ---------- main ----------
func main() {
	cfg := loadConfig()

	words, err := loadEmbeddedWordlist()
	if err != nil {
		log.Fatal(err)
	}
	svc, err := newService(words)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting %s", cfg.ServerName)
	log.Printf("repo: %s", cfg.RepoURL)
	log.Printf("image: %s", cfg.GHCRImage)
	log.Printf("version=%s commit=%s", version, commit)

	mux := http.NewServeMux()

	// https://github.com/sparkes/fixphrase-service/tree/main?tab=readme-ov-file#get
	// GET /encode?lat=..&lon=..
	// POST /encode { "lat": .., "lon": .. }
	mux.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		var lat, lon float64

		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			latS := q.Get("lat")
			lonS := q.Get("lon")
			if latS == "" || lonS == "" {
				writeJSON(w, 400, map[string]any{
					"error": "missing query params: lat and lon are required",
					"usage": []string{
						"GET  /encode?lat=52.5902&lon=-2.13049",
						`POST /encode {"lat":52.5902,"lon":-2.13049}`,
					},
				})
				return
			}
			var err error
			lat, err = strconv.ParseFloat(latS, 64)
			if err != nil {
				writeJSON(w, 400, map[string]any{"error": "invalid lat", "detail": err.Error()})
				return
			}
			lon, err = strconv.ParseFloat(lonS, 64)
			if err != nil {
				writeJSON(w, 400, map[string]any{"error": "invalid lon", "detail": err.Error()})
				return
			}

		case http.MethodPost:
			var req EncodeRequest
			if err := readJSON(r, &req); err != nil {
				writeJSON(w, 400, map[string]any{"error": "invalid json", "detail": err.Error()})
				return
			}
			lat, lon = req.Lat, req.Lon

		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}

		res, err := encodeFixPhrase(svc, lat, lon)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, 200, res)
	})

	// https://github.com/sparkes/fixphrase-service/tree/main?tab=readme-ov-file#get-1
	// GET /decode?phrase=word1%20word2%20word3%20word4
	// POST /decode { "phrase": "word1 word2 ..." }
	mux.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
		var phrase string

		switch r.Method {
		case http.MethodGet:
			phrase = r.URL.Query().Get("phrase")
			if strings.TrimSpace(phrase) == "" {
				writeJSON(w, 400, map[string]any{
					"error": "missing query param: phrase",
					"usage": []string{
						"GET  /decode?phrase=abacus%20abdomen%20...",
						`POST /decode {"phrase":"abacus abdomen ..."}`,
					},
				})
				return
			}
		case http.MethodPost:
			var req DecodeRequest
			if err := readJSON(r, &req); err != nil {
				writeJSON(w, 400, map[string]any{"error": "invalid json", "detail": err.Error()})
				return
			}
			phrase = req.Phrase
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}

		res, err := decodeFixPhrase(svc, phrase)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}

		// Optional nicety: also show sorted input words.
		sorted := append([]string{}, res.InputWords...)
		sort.Strings(sorted)

		writeJSON(w, 200, map[string]any{
			"result":           res,
			"inputWordsSorted": sorted,
		})
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"ok":          true,
			"service":     cfg.ServerName,
			"repo":        cfg.RepoURL,
			"image":       cfg.GHCRImage,
			"version":     version,
			"commit":      commit,
			"buildDate":   date,
			"wordlistLen": len(svc.words),
			"time":        time.Now().UTC().Format(time.RFC3339),
		})
	})

	addr := cfg.Addr

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("listening on %s (version=%s commit=%s)", addr, version, commit)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Printf("shutdown requested")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = server.Close()
	}
	log.Printf("shutdown complete")
}
