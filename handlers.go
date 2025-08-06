package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/m-kowalsky/weigh-in/internal/database"
	"github.com/markbates/goth/gothic"
)

const sess_email = "user_email"

const sess_userId = "user_id"

type PageData struct {
	User        database.User
	Provider    string
	Title       string
	ChartHTML   template.HTML
	CurrentDate string
}

func (cfg *apiConfig) handlerKamalHealthcheck(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerGetAuthCallback(w http.ResponseWriter, r *http.Request) {

	// Cookie debug
	// for _, cookie := range r.Cookies() {
	// 	fmt.Printf("\nGetAuthCallback-start cookie: name: %v, value: %v\n", cookie.Name, cookie.Value)
	// }

	// Get provider param from url for gothic auth
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	goth_user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println("Auth error:", err)
		fmt.Fprintln(w, err)
		return
	}

	// Cookie debug
	// for _, cookie := range r.Cookies() {
	// 	fmt.Printf("\nGetAuthCallback-end cookie: name: %v, value: %v\n", cookie.Name, cookie.Value)
	// }

	// Check if user exists in db already by getting a count of a user by email
	count, err := cfg.db.CheckIfUserExistsByEmail(r.Context(), goth_user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if count == 1 {
		current_user, err := cfg.db.GetUserByEmail(r.Context(), goth_user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("current user: %v", current_user)
	} else {
		new_user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Email:       goth_user.Email,
			AccessToken: goth_user.AccessToken,
			FullName:    sql.NullString{String: goth_user.Name, Valid: true},
			Provider:    provider,
		})
		if err != nil {
			http.Error(w, "problem creating new user", http.StatusBadRequest)
		}

		fmt.Printf("current user: %v", new_user)
	}

	// Create new session with user id and email

	sess, err := gothic.Store.Get(r, session_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sess.Values[sess_email] = goth_user.Email
	sess.Values[sess_userId] = goth_user.UserID
	sess.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (cfg *apiConfig) handlerGetAuth(w http.ResponseWriter, r *http.Request) {

	// Cookie debug
	// for _, cookie := range r.Cookies() {
	// 	fmt.Printf("\nGetAuth cookie: name: %v, value: %v\n", cookie.Name, cookie.Value)
	// }

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		tmpl.ExecuteTemplate(w, "index", gothUser)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (cfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {

	// Logout a user by setting the current user_session max age to -1 which will cause the client to delete the cookie associated with the session
	session, err := gothic.Store.Get(r, session_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (cfg *apiConfig) handlerIndex(w http.ResponseWriter, r *http.Request) {

	current_user, err := cfg.getCurrentUser(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	chart_data, err := cfg.getChartData(w, r)
	if err != nil {
		log.Fatal("Failed to get chart data in index handler")
	}

	current_date := time.Now().Format("2006-01-02")

	data := PageData{
		User:        current_user,
		Provider:    current_user.Provider.(string),
		Title:       "Weigh In",
		ChartHTML:   template.HTML(chart_data),
		CurrentDate: current_date,
	}

	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	user_id := chi.URLParam(r, "user_id")

	id_int, err := strconv.ParseInt(user_id, 16, 64)
	if err != nil {
		http.Error(w, "Faile to convert user_id urlParam to int", http.StatusBadRequest)
	}

	current_user, err := cfg.db.GetUserById(r.Context(), id_int)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("user email from handerlGetUser before tmpl execute: %v\n", current_user.Email)

	err = tmpl.ExecuteTemplate(w, "user", current_user)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handlerWeighInNew(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "weigh_in_form", nil)

}

func (cfg *apiConfig) handlerLandingPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "landing_page", nil)

}

func (cfg *apiConfig) handlerCreateWeighIn(w http.ResponseWriter, r *http.Request) {

	current_user, err := cfg.getCurrentUser(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	// convert weight string to int64
	weight, err := strconv.ParseInt(r.FormValue("weight"), 10, 64)
	if err != nil {
		http.Error(w, "Failed to parse weight to int64", http.StatusBadRequest)
		return
	}

	// convert cheated and alcohol string to bool
	cheated := false
	alcohol := false
	if r.FormValue("cheated") == "on" {
		cheated = true
	}
	if r.FormValue("alcohol") == "on" {
		alcohol = true
	}

	// convert log date string to time.Time
	time_layout := "2006-01-02"
	log_date, err := time.Parse(time_layout, r.FormValue("log_date"))

	weighInNew, err := cfg.db.CreateWeighIn(r.Context(), database.CreateWeighInParams{
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Weight:      weight,
		WeightUnit:  r.FormValue("weight_unit"),
		LogDate:     log_date,
		Cheated:     cheated,
		Alcohol:     alcohol,
		Note:        sql.NullString{String: r.FormValue("note"), Valid: true},
		WeighInDiet: r.FormValue("weigh_in_diet"),
		UserID:      current_user.ID,
	})
	if err != nil {
		http.Error(w, "Failed to create new weigh in", http.StatusBadRequest)
		return
	}
	fmt.Printf("weigh in: %+v\n", weighInNew)

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		ProviderIndex *ProviderIndex
		Title         string
	}

	data := LoginData{
		ProviderIndex: cfg.providerIndex,
		Title:         "Weigh In - Login",
	}
	err := tmpl.ExecuteTemplate(w, "login", data)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handlerRefreshChart(w http.ResponseWriter, r *http.Request) {

	chart_html, err := cfg.getChartData(w, r)
	if err != nil {
		log.Fatal("Failed to get chart data in refresh chart handler")
	}
	data := PageData{
		ChartHTML: template.HTML(chart_html),
	}

	err = tmpl.ExecuteTemplate(w, "line_chart", data)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) getCurrentUser(w http.ResponseWriter, r *http.Request) (database.User, error) {
	sess, _ := gothic.Store.Get(r, session_name)
	if sess.IsNew == true {
		fmt.Println(sess.Options.MaxAge)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return database.User{}, nil
	}
	user_email := sess.Values["user_email"]

	current_user, err := cfg.db.GetUserByEmail(r.Context(), user_email.(string))
	if err != nil {
		return database.User{}, err
	}
	return current_user, nil
}

type ChartData struct {
	XAxis    []string
	LineData []opts.LineData
}

func (cfg *apiConfig) getChartData(w http.ResponseWriter, r *http.Request) ([]byte, error) {

	user, err := cfg.getCurrentUser(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	// Get current date and subtract 30 days to get date range for chart
	current_date := time.Now()

	date_range_min := current_date.AddDate(0, 0, -30)

	weighIn_data, err := cfg.db.GetWeightChartDataByUser(r.Context(), database.GetWeightChartDataByUserParams{
		UserID:  user.ID,
		LogDate: date_range_min,
	})

	if err != nil {
		log.Fatal("Failed to get weighIn data in getChartData()")
	}

	chart_data := ChartData{}
	for _, weighIn := range weighIn_data {
		chart_data.XAxis = append(chart_data.XAxis, weighIn.LogDate.Format("01-02"))
		chart_data.LineData = append(chart_data.LineData, opts.LineData{Value: weighIn.Weight})
	}
	chart_html := renderChartContent(chart_data)

	return chart_html, nil
}

func renderChartContent(data ChartData) []byte {

	chart := charts.NewLine()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Weigh Ins", Subtitle: "Last 30 days"}))

	chart.SetXAxis(data.XAxis).AddSeries("Weight", data.LineData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Show: opts.Bool(true),
		}),
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: opts.Float(0.2),
		}),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
	)
	chartHTML := chart.RenderContent()

	return chartHTML
}
