package akinator

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AkinatorResponse struct {
	AkinatorGuess
	AkinatorNextQuestion
}

type AkinatorNextQuestion struct {
	Completion  string `json:"completion"`
	Akitude     string `json:"akitude"`
	Step        string `json:"step"`
	Progression string `json:"progression"`
	QuestionID  string `json:"question_id"`
	Question    string `json:"question"`
}

type AkinatorGuess struct {
	DescriptionProposition string `json:"description_proposition"`
	FlagPhoto              string `json:"flag_photo"`
	IDBaseProposition      string `json:"id_base_proposition"`
	IDProposition          string `json:"id_proposition"`
	NameProposition        string `json:"name_proposition"`
	NbElements             int    `json:"nb_elements"`
	Photo                  string `json:"photo"`
	Pseudo                 string `json:"pseudo"`
	ValideConstrainte      string `json:"valide_constrainte"`
}

type Akinator struct {
	uri             string
	httpClient      http.Client
	Session         string
	Signature       string
	Answers         []string
	CurrentQuestion string
	Progress        float64
	CurrentStep     int
	Guess           AkinatorGuess
}

func NewAkinator(language string) *Akinator {
	uri := fmt.Sprintf("https://%s.akinator.com", language)
	return &Akinator{
		uri: uri,
		httpClient: http.Client{
			Timeout: time.Second * 30,
		},
		Answers: make([]string, 0),
	}
}

func (a *Akinator) Start() error {
	form := url.Values{}
	form.Add("sid", "1")
	form.Add("cm", "false")

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/game", a.uri), strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("unknown error occurred")
		os.Exit(1)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := a.httpClient.Do(request)
	if err != nil {
		fmt.Println("unknown error occurred")
		os.Exit(1)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("unknown error occurred")
		os.Exit(1)
	}

	text := string(responseBytes)

	question := firstMatch(text, `<p class="question-text" id="question-label">(.+?)<\/p>`)
	session := firstMatch(text, `session: '(.+)'`)
	signature := firstMatch(text, "signature: '(.+)'")

	a.Session = session
	a.Signature = signature
	a.CurrentQuestion = question
	a.CurrentStep = 0

	a.Answers = append(a.Answers,
		html.UnescapeString(firstMatch(text, `<a class="li-game" href="#" id="a_yes" onclick="chooseAnswer\(0\)">(.+)<\/a>`)),         // Yes
		html.UnescapeString(firstMatch(text, `<a class="li-game" href="#" id="a_no" onclick="chooseAnswer\(1\)">(.+)<\/a>`)),          // No
		html.UnescapeString(firstMatch(text, `<a class="li-game" href="#" id="a_dont_know" onclick="chooseAnswer\(2\)">(.+)<\/a>`)),   // Dont Know
		html.UnescapeString(firstMatch(text, `<a class="li-game" href="#" id="a_probably" onclick="chooseAnswer\(3\)">(.+)<\/a>`)),    // Probably
		html.UnescapeString(firstMatch(text, `<a class="li-game" href="#" id="a_probaly_not" onclick="chooseAnswer\(4\)">(.+)<\/a>`)), // Probably Not
	)

	return nil
}

// Continue after a wrong anwser
func (a *Akinator) KeepGuessing() (err error) {
	form := url.Values{}
	form.Add("step", fmt.Sprintf("%d", a.CurrentStep))
	form.Add("progression", fmt.Sprintf("%.2f", a.Progress))
	form.Add("sid", "1")
	form.Add("cm", "false")
	form.Add("session", a.Session)
	form.Add("signature", a.Signature)

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/exclude", a.uri), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := a.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var akinatorResponse AkinatorResponse
	err = json.NewDecoder(response.Body).Decode(&akinatorResponse)
	if err != nil {
		return err
	}

	a.CurrentStep, err = strconv.Atoi(akinatorResponse.Step)
	if err != nil {
		return err
	}
	a.Progress, err = strconv.ParseFloat(akinatorResponse.Progression, 32)
	if err != nil {
		return err
	}
	a.CurrentQuestion = akinatorResponse.Question

	return nil
}

func (a *Akinator) NextQuestion(answer int) (guess AkinatorGuess, err error) {
	form := url.Values{}

	form.Add("step", fmt.Sprintf("%d", a.CurrentStep))
	form.Add("progression", fmt.Sprintf("%.2f", a.Progress))
	form.Add("sid", "1")
	form.Add("cm", "false")
	form.Add("answer", fmt.Sprintf("%d", answer))
	form.Add("step_last_proposition", "")
	form.Add("session", a.Session)
	form.Add("signature", a.Signature)

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/answer", a.uri), strings.NewReader(form.Encode()))
	if err != nil {
		return guess, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := a.httpClient.Do(request)
	if err != nil {
		return guess, err
	}
	defer response.Body.Close()

	var akinatorResponse AkinatorResponse
	//err = json.NewDecoder(response.Body).Decode(&akinatorResponse)
	//if err != nil {
	//	return guess, err
	//}

	bytesResponse, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytesResponse, &akinatorResponse)
	if err != nil {
		panic(err)
	}

	// Found the character
	if akinatorResponse.IDProposition != "" {
		a.Guess = akinatorResponse.AkinatorGuess
		return a.Guess, nil
	}

	a.CurrentStep, err = strconv.Atoi(akinatorResponse.Step)
	if err != nil {
		return guess, err
	}
	a.Progress, err = strconv.ParseFloat(akinatorResponse.Progression, 32)
	if err != nil {
		return guess, err
	}
	a.CurrentQuestion = akinatorResponse.Question

	return guess, nil
}

func (a *Akinator) Back() (err error) {
	form := url.Values{}
	form.Add("step", fmt.Sprintf("%d", a.CurrentStep))
	form.Add("progression", fmt.Sprintf("%.2f", a.Progress))
	form.Add("sid", "1")
	form.Add("cm", "false")
	form.Add("session", a.Session)
	form.Add("signature", a.Signature)

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/cancel_answer", a.uri), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := a.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var akinatorResponse AkinatorResponse
	err = json.NewDecoder(response.Body).Decode(&akinatorResponse)
	if err != nil {
		return err
	}

	a.CurrentStep, err = strconv.Atoi(akinatorResponse.Step)
	if err != nil {
		return err
	}
	a.Progress, err = strconv.ParseFloat(akinatorResponse.Progression, 32)
	if err != nil {
		return err
	}
	a.CurrentQuestion = akinatorResponse.Question

	return nil
}

func firstMatch(text string, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)

	match := ""
	if len(matches) > 1 {
		match = matches[1]
	}

	return match
}
