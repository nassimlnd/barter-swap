package main

import "net/http"

type createUserRequest struct {
	Pseudo string `json:"pseudo"`
	Bio    string `json:"bio"`
	Ville  string `json:"ville"`
}

type updateUserRequest struct {
	Pseudo string `json:"pseudo"`
	Bio    string `json:"bio"`
	Ville  string `json:"ville"`
}

type replaceSkillsRequest struct {
	Skills []skillDTO `json:"skills"`
}

func (a *App) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	u, err := a.Users.Create(r.Context(), CreateUserInput{Pseudo: req.Pseudo, Bio: req.Bio, Ville: req.Ville})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, userToDTO(u))
}

func (a *App) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	u, err := a.Users.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userToDTO(u))
}

func (a *App) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req updateUserRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	u, err := a.Users.Update(r.Context(), callerID, id, UpdateUserInput{Pseudo: req.Pseudo, Bio: req.Bio, Ville: req.Ville})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userToDTO(u))
}

func (a *App) handleGetUserSkills(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	skills, err := a.Users.ListSkills(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skillsToDTO(skills))
}

func (a *App) handleReplaceUserSkills(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req replaceSkillsRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	skills := make([]Skill, 0, len(req.Skills))
	for _, s := range req.Skills {
		skills = append(skills, Skill{Nom: s.Nom, Niveau: SkillNiveau(s.Niveau)})
	}
	out, err := a.Users.ReplaceSkills(r.Context(), callerID, id, skills)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skillsToDTO(out))
}
