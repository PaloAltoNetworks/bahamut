package {{ handlers_package_name }}

import (
  "encoding/json"
  "net/http"

    log "github.com/Sirupsen/logrus"
    "github.com/aporeto-inc/bahamut"
    "github.com/aporeto-inc/elemental"

    "{{ models_package_package }}"
)

// RetrieveMany{{ specification.entity_name }} handles GET requests for a set of {{ specification.entity_name }}.
func RetrieveMany{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationRetrieveMany)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling retrieve many {{ specification.entity_name|lower }} request.")

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveManyProcessor); !ok {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "No handler for retrieving many {{ specification.resource_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    ctx.WriteResponse(w)
}

// Retrieve{{ specification.entity_name }} handles GET requests for a single {{ specification.entity_name }}.
func Retrieve{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationRetrieve)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling retrieve {{ specification.entity_name|lower }} request.")

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveProcessor); !ok {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "No handler for retrieving a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    ctx.WriteResponse(w)
}

// Create{{ specification.entity_name }} handles POST requests for a single {{ specification.entity_name }}.
func Create{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationCreate)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling create {{ specification.entity_name|lower }} request.")

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.CreateProcessor); !ok {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    obj := {{ models_package_name }}.New{{ specification.entity_name }}()
    if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    if err := obj.Validate(); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    ctx.InputData = obj

    if err := proc.(bahamut.CreateProcessor).ProcessCreate(ctx); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    if !ctx.HasErrors() {

        if ctx.HasEvents() {
            server.Push(ctx.Events()...)
        }

        if ctx.OutputData != nil {
            server.Push(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*{{ models_package_name }}.{{ specification.entity_name }})))
        }
    }

    ctx.WriteResponse(w)
}

// Update{{ specification.entity_name }} handles PUT requests for a single {{ specification.entity_name }}.
func Update{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationUpdate)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling update {{ specification.entity_name|lower }} request.")

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.UpdateProcessor); !ok {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "No handler for updating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    obj := {{ models_package_name }}.New{{ specification.entity_name }}()
    if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "http", http.StatusBadRequest))
        return
    }

    if err := obj.Validate(); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    ctx.InputData = obj

    if err := proc.(bahamut.UpdateProcessor).ProcessUpdate(ctx); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    if !ctx.HasErrors() {

        if ctx.HasEvents() {
            server.Push(ctx.Events()...)
        }

        if ctx.OutputData != nil {
            server.Push(elemental.NewEvent(elemental.EventUpdate, ctx.OutputData.(*{{ models_package_name }}.{{ specification.entity_name }})))
        }
    }

    ctx.WriteResponse(w)
}

// Delete{{ specification.entity_name }} handles DELETE requests for a single {{ specification.entity_name }}.
func Delete{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationDelete)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling delete {{ specification.entity_name|lower }} request.")

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveProcessor); !ok {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "No handler for retrieving a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.DeleteProcessor).ProcessDelete(ctx); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    if !ctx.HasErrors() {

        if ctx.HasEvents() {
          server.Push(ctx.Events()...)
        }

        if ctx.OutputData != nil {
          server.Push(elemental.NewEvent(elemental.EventDelete, ctx.OutputData.(*{{ models_package_name }}.{{ specification.entity_name }})))
        }
    }

    ctx.WriteResponse(w)
}

// Patch{{ specification.entity_name }} handles PATCH requests for a single {{ specification.entity_name }}.
func Patch{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationPatch)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling patch {{ specification.entity_name|lower }} request.")

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.PatchProcessor); !ok {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "No handler for patching a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    var assignation *elemental.Assignation
    if err := json.NewDecoder(req.Body).Decode(&assignation); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "http", http.StatusBadRequest))
        return
    }

    ctx.InputData = assignation

    if err := proc.(bahamut.PatchProcessor).ProcessPatch(ctx); err != nil {
        bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), err)
        return
    }

    if !ctx.HasErrors() {

        if ctx.HasEvents() {
            server.Push(ctx.Events()...)
        }

        if ctx.OutputData != nil {
            server.Push(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*elemental.Assignation)))
        }
    }

    ctx.WriteResponse(w)
}

// Info{{ specification.entity_name }} handles HEAD requests for a single {{ specification.entity_name }}.
func Info{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    ctx := bahamut.NewContext(elemental.OperationInfo)
    ctx.ReadRequest(req)

    log.WithFields(log.Fields{
        "package":    "bahamut",
        "origin":     req.RemoteAddr,
        "context":    ctx.String(),
    }).Debug("Handling info {{ specification.entity_name|lower }} request.")

    bahamut.WriteHTTPError(w, ctx.Info.Headers.Get("Origin"), elemental.NewError("Not implemented", "Info{{ specification.entity_name }} not implemented in Cid yet", "http", http.StatusNotImplemented))
}
