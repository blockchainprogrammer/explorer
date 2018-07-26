package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gochain-io/explorer/api/backend"
	"github.com/gochain-io/explorer/api/models"

	"github.com/codegangsta/cli"
	"github.com/gochain-io/gochain/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

var ethClient *ethclient.Client
var mongoBackend *backend.MongoBackend

func getClient(url string) *ethclient.Client {
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Fatal().Err(err).Msg("main")
	}
	return client
}

func main() {
	var url string
	var loglevel string
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "rpc-url, u",
			Value:       "https://rpc.gochain.io",
			Usage:       "rpc api url, 'https://rpc.gochain.io'",
			Destination: &url,
		},
		cli.StringFlag{
			Name:        "log, l",
			Value:       "info",
			Usage:       "loglevel debug/info/warn/fatal, default is Info",
			Destination: &loglevel,
		},
	}

	app.Action = func(c *cli.Context) error {
		level, _ := zerolog.ParseLevel(loglevel)
		zerolog.SetGlobalLevel(level)
		client := getClient(url)
		mongoBackend = backend.NewBackend(client)
		ethClient = client
		r := chi.NewRouter()
		// A good base middleware stack
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		cors2 := cors.New(cors.Options{
			// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origin"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})
		r.Use(cors2.Handler)
		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		r.Use(middleware.Timeout(60 * time.Second))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome"))
		})
		r.Route("/blocks", func(r chi.Router) {
			r.Get("/", listBlocks)
			r.Get("/{num}", getBlock)
		})
		http.ListenAndServe(":8080", r)
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("Run")
	}

}

func listBlocks(w http.ResponseWriter, r *http.Request) {
	bl := &models.BlockList{
		Blocks: []*models.Block{},
	}
	bl.Blocks = mongoBackend.GetLatestsBlocks(10)
	writeJSON(w, http.StatusOK, bl)
}

func getBlock(w http.ResponseWriter, r *http.Request) {
	bnumS := chi.URLParam(r, "num")
	bnum, err := strconv.Atoi(bnumS)
	if err != nil {
		log.Error().Err(err).Str("bnumS", bnumS).Msg("Error converting bnumS to num")
		// todo: sendError()
		return
	}
	log.Info().Int("bnum", bnum).Msg("looking up block")
	block := mongoBackend.GetBlockByNumber(int64(bnum))
	writeJSON(w, http.StatusOK, block)
}
