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

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  bahamut.OperationRetrieveMany,
        "path":       req.URL.Path,
    }).Debug("handling retrieve many {{ specification.entity_name|lower }} request")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(bahamut.OperationRetrieveMany)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity(models.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveManyProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
        return
    }

    ctx.WriteResponse(w)
}

// Retrieve{{ specification.entity_name }} handles GET requests for a single {{ specification.entity_name }}.
func Retrieve{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  bahamut.OperationRetrieve,
        "path":       req.URL.Path,
    }).Debug("handling retrieve {{ specification.entity_name|lower }} request")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(bahamut.OperationRetrieve)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity(models.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
        return
    }

    ctx.WriteResponse(w)
}

// Create{{ specification.entity_name }} handles POST requests for a single {{ specification.entity_name }}.
func Create{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    log.WithFields(log.Fields{
      "context":    "handler",
      "origin":     req.RemoteAddr,
      "method":     req.Method,
      "operation":  bahamut.OperationCreate,
      "path":       req.URL.Path,
    }).Debug("handling create request")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(bahamut.OperationCreate)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity(models.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.CreateProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    var obj *models.{{ specification.entity_name }}
    if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
        bahamut.WriteHTTPError(w, http.StatusBadRequest, elemental.NewError("Bad Request", "The request cannot be processed", "http", http.StatusBadRequest))
        return
    }

    ctx.AddErrors(obj.Validate()...)
    if ctx.HasErrors() {
        bahamut.WriteHTTPError(w, http.StatusConflict, ctx.Errors...)
        return
    }

    ctx.InputData = obj

    if err := proc.(bahamut.CreateProcessor).ProcessCreate(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
        return
    }

    if !ctx.HasErrors() {

        if len(ctx.EventsQueue) > 0 {
            server.Push(ctx.EventsQueue[0])
        }

        if ctx.OutputData != nil {
            server.Push(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*models.{{ specification.entity_name }})))
        }
    }

    ctx.WriteResponse(w)
}

// Update{{ specification.entity_name }} handles PUT requests for a single {{ specification.entity_name }}.
func Update{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  bahamut.OperationUpdate,
        "path":       req.URL.Path,
    }).Debug("handling update {{ specification.entity_name|lower }} request")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(bahamut.OperationUpdate)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity(models.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    var obj *models.{{ specification.entity_name }}
    if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
        bahamut.WriteHTTPError(w, http.StatusBadRequest, elemental.NewError("Bad Request", "The request cannot be processed", "http", http.StatusBadRequest))
        return
    }

    ctx.AddErrors(obj.Validate()...)
    if ctx.HasErrors() {
        bahamut.WriteHTTPError(w, http.StatusConflict, ctx.Errors...)
        return
    }

    ctx.InputData = obj

    if err := proc.(bahamut.UpdateProcessor).ProcessUpdate(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
        return
    }

    if !ctx.HasErrors() {

        if len(ctx.EventsQueue) > 0 {
            server.Push(ctx.EventsQueue...)
        }

        if ctx.OutputData != nil {
            server.Push(elemental.NewEvent(elemental.EventUpdate, ctx.OutputData.(*models.{{ specification.entity_name }})))
        }
    }

    ctx.WriteResponse(w)
}

// Delete{{ specification.entity_name }} handles DELETE requests for a single {{ specification.entity_name }}.
func Delete{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  bahamut.OperationDelete,
        "path":       req.URL.Path,
    }).Debug("handling delete {{ specification.entity_name|lower }} request")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(bahamut.OperationDelete)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity(models.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.DeleteProcessor).ProcessDelete(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
        return
    }

    if !ctx.HasErrors() {

        if len(ctx.EventsQueue) > 0 {
          server.Push(ctx.EventsQueue...)
        }

        if ctx.OutputData != nil {
          server.Push(elemental.NewEvent(elemental.EventDelete, ctx.OutputData.(*models.{{ specification.entity_name }})))
        }
    }

    ctx.WriteResponse(w)
}

// Patch{{ specification.entity_name }} handles PATCH requests for a single {{ specification.entity_name }}.
func Patch{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  bahamut.OperationPatch,
        "path":       req.URL.Path,
    }).Debug("handling patch {{ specification.entity_name|lower }} request")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(bahamut.OperationPatch)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity(models.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.PatchProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for patching a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    var assignation *elemental.Assignation
    if err := json.NewDecoder(req.Body).Decode(&assignation); err != nil {
        bahamut.WriteHTTPError(w, http.StatusBadRequest, elemental.NewError("Bad Request", "The request cannot be processed", "http", http.StatusBadRequest))
        return
    }

    ctx.InputData = assignation

    if err := proc.(bahamut.PatchProcessor).ProcessPatch(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
        return
    }

    if !ctx.HasErrors() {

        if len(ctx.EventsQueue) > 0 {
            server.Push(ctx.EventsQueue[0])
        }

        if ctx.OutputData != nil {
            server.Push(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*elemental.Assignation)))
        }
    }

    ctx.WriteResponse(w)
}

// Info{{ specification.entity_name }} handles HEAD requests for a single {{ specification.entity_name }}.
func Info{{ specification.entity_name }}(w http.ResponseWriter, req *http.Request) {

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  bahamut.OperationInfo,
        "path":       req.URL.Path,
    }).Debug("handling info {{ specification.entity_name|lower }} request")

    bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "Info{{ specification.entity_name }} not implemented in Cid yet", "http", http.StatusNotImplemented))
}
