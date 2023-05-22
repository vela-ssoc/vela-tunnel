package tunnel

import "net/http"

type webapp struct {
	code Coder
	proc Processor
}

func (app *webapp) substance(w http.ResponseWriter, r *http.Request) {
	var req substanceRequest
	err := app.bindJSON(r, &req)
	if err != nil || len(req.Removes) == 0 || len(req.Updates) == 0 {
		return
	}

	ctx := r.Context()
	dats, err := app.proc.Substance(ctx, req.Removes, req.Updates)
	res := &dataResponse{Data: dats}

	_ = app.writeJSON(w, res)
}

func (app *webapp) third(w http.ResponseWriter, r *http.Request) {
	var req thirdRequest
	err := app.bindJSON(r, &req)
	id := req.ID
	if err != nil || id == 0 || req.Event == "" {
		return
	}

	ctx := r.Context()
	switch req.Event {
	case "update":
		err = app.proc.ThirdUpdate(ctx, id)
	case "remove":
		err = app.proc.ThirdRemove(ctx, id)
	}

	w.WriteHeader(http.StatusOK)
}

func (app *webapp) bindJSON(r *http.Request, v any) error {
	return app.code.NewDecoder(r.Body).Decode(v)
}

func (app *webapp) writeJSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return app.code.NewEncoder(w).Encode(v)
}

type thirdRequest struct {
	ID    int64  `json:"id"`
	Event string `json:"event"`
}

type substanceRequest struct {
	Removes []int64      `json:"removes"`
	Updates []*TaskChunk `json:"updates"`
}

type dataResponse struct {
	Data any `json:"data"`
}
