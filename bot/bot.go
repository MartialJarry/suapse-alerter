package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"

	"github.com/bwmarrin/discordgo" //discordgo package from the repo of bwmarrin .
)

type configStruct struct {
	Token    string
	Username string
	Password string
}

type jsonStruct struct {
	Code             string `json:"code"`
	Civility         string `json:"civility"`
	Name             string `json:"name"`
	Firstname        string `json:"firstname"`
	Email            string `json:"email"`
	BirthDate        string `json:"birthDate"`
	EstBoursier      bool   `json:"estBoursier"`
	Composante       string `json:"composante"`
	Departement      string `json:"departement"`
	TypePersonne     string `json:"typePersonne"`
	MontantPaiement  int    `json:"montantPaiement"`
	PaiementEffectue bool   `json:"paiementEffectue"`
	Sports           []struct {
		Code      int `json:"code"`
		Categorie struct {
			Code    int    `json:"code"`
			Nom     string `json:"nom"`
			Picto   string `json:"picto"`
			Image   string `json:"image"`
			Couleur string `json:"couleur"`
		} `json:"categorie"`
		Registrations []interface{} `json:"registrations"`
		Creneaux      []struct {
			Site            string `json:"site"`
			Places          int    `json:"places"`
			Code            int    `json:"code"`
			Jour            string `json:"jour"`
			Encadrant       string `json:"encadrant"`
			Heures          string `json:"heures"`
			Localisation    string `json:"localisation"`
			Adresse         string `json:"adresse"`
			Niveau          string `json:"niveau"`
			PlacesRestantes int    `json:"placesRestantes"`
		} `json:"creneaux"`
		Description string `json:"description"`
		Nom         string `json:"nom"`
	} `json:"sports"`
}

var (
	Config     configStruct
	BotId      string
	goBot      *discordgo.Session
	prevPlaces int
	debug      bool
	botChannel string
	jsondata   jsonStruct
)

func Start() {
	err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	goBot, err := discordgo.New("Bot " + Config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Making our bot a user using User function .
	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	BotId = u.ID

	goBot.AddHandler(messageHandler)

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running !")
	pollSuapse(goBot)
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotId {
		return
	}
	fmt.Println("Message reçu !", m.Content, m.ChannelID)

	if m.Content == "volley" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "ball")
		if botChannel != m.ChannelID {
			botChannel = m.ChannelID
			fmt.Println("Channel set ! Bot will post on channel #id", botChannel)
			_, _ = s.ChannelMessageSend(m.ChannelID, "channel enregistré !")
		}
	}
}

func readConfig() error {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	Config.Token = os.Getenv("TOKEN")
	Config.Username = os.Getenv("USERNAME")
	Config.Password = os.Getenv("PASSWORD")

	return nil
}

func pollSuapse(s *discordgo.Session) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("https://unsport.univ-nantes.fr/web/api/user").MustWaitLoad()

	wait := page.MustWaitRequestIdle()

	text := page.MustElement("body").MustText()

	page.MustElement("#username").MustInput(Config.Username)
	page.MustElement("#password").MustInput(Config.Password).MustPress(input.Enter)

	wait()

	for {
		page := browser.MustPage("https://unsport.univ-nantes.fr/web/api/user").MustWaitLoad()
		text = page.MustElement("body").MustText()
		if debug == true {
			fmt.Println(text)
		}

		err := json.Unmarshal([]byte(text), &jsondata)
		if err != nil {
			fmt.Println(err)
		}
		//creneau := jsondata.Sports[1].Creneaux[4]
		for _, sport := range jsondata.Sports {
			//fmt.Println(i,sport)
			if sport.Nom == "Volley ball" {
				for _, creneau := range sport.Creneaux {
					//fmt.Println(i2,creneau)
					if creneau.Code == 933 {
						msg := fmt.Sprintf("Lieu : %s, places restantes : %d", creneau.Localisation, creneau.PlacesRestantes)
						fmt.Println(msg)

						if creneau.PlacesRestantes > prevPlaces && prevPlaces == 0 {
							msg := "@everyone PLACES REMPLIES - ALLEZ RESERVER"
							fmt.Println(msg)
							if botChannel != "" {
								_, _ = s.ChannelMessageSend(botChannel, msg)
							}
						} else if creneau.PlacesRestantes < prevPlaces {
							msg := fmt.Sprintf("@here Plus que %d places disponibles", creneau.PlacesRestantes)
							fmt.Println(msg)
							if botChannel != "" {
								_, _ = s.ChannelMessageSend(botChannel, msg)
							}
						}

						prevPlaces = creneau.PlacesRestantes
					}
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
