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
        "operation":  elemental.OperationRetrieveMany,
        "path":       req.URL.Path,
    }).Debug("Handling retrieve many {{ specification.entity_name|lower }} request.")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationRetrieveMany)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

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
        "operation":  elemental.OperationRetrieve,
        "path":       req.URL.Path,
    }).Debug("Handling retrieve {{ specification.entity_name|lower }} request.")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationRetrieve)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

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
      "operation":  elemental.OperationCreate,
      "path":       req.URL.Path,
    }).Debug("Handling create {{ specification.entity_name|lower }} request.")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationCreate)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.CreateProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    var obj *{{ models_package_name }}.{{ specification.entity_name }}
    if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
        bahamut.WriteHTTPError(w, http.StatusBadRequest, elemental.NewError("Bad Request", "The request cannot be processed", "http", http.StatusBadRequest))
        return
    }

    if errs := obj.Validate(); errs != nil {
        bahamut.WriteHTTPError(w, http.StatusExpectationFailed, errs...)
        return
    }

    ctx.InputData = obj

    if err := proc.(bahamut.CreateProcessor).ProcessCreate(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
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

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  elemental.OperationUpdate,
        "path":       req.URL.Path,
    }).Debug("Handling update {{ specification.entity_name|lower }} request.")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationUpdate)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.UpdateProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    defer req.Body.Close()
    var obj *{{ models_package_name }}.{{ specification.entity_name }}
    if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
        bahamut.WriteHTTPError(w, http.StatusBadRequest, elemental.NewError("Bad Request", "The request cannot be processed", "http", http.StatusBadRequest))
        return
    }

    if errs := obj.Validate(); errs != nil {
        bahamut.WriteHTTPError(w, http.StatusExpectationFailed, errs...)
        return
    }

    ctx.InputData = obj

    if err := proc.(bahamut.UpdateProcessor).ProcessUpdate(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
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

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  elemental.OperationDelete,
        "path":       req.URL.Path,
    }).Debug("Handling delete {{ specification.entity_name|lower }} request.")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationDelete)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

    if _, ok := proc.(bahamut.RetrieveProcessor); !ok {
        bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "No handler for creating a {{ specification.rest_name }}", "http", http.StatusNotImplemented))
        return
    }

    if err := proc.(bahamut.DeleteProcessor).ProcessDelete(ctx); err != nil {
        bahamut.WriteHTTPError(w, http.StatusInternalServerError, elemental.NewError("Internal Server Error", err.Error(), "http", http.StatusInternalServerError))
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

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  elemental.OperationPatch,
        "path":       req.URL.Path,
    }).Debug("Handling patch {{ specification.entity_name|lower }} request.")

    server := bahamut.DefaultBahamut()
    ctx := bahamut.NewContext(elemental.OperationPatch)
    ctx.ReadRequest(req)

    if !bahamut.CheckAuthentication(ctx, w) {
        return
    }

    if !bahamut.CheckAuthorization(ctx, w) {
        return
    }

    proc, _ := server.ProcessorForIdentity({{ models_package_name }}.{{ specification.entity_name }}Identity)

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

    log.WithFields(log.Fields{
        "context":    "handler",
        "origin":     req.RemoteAddr,
        "method":     req.Method,
        "operation":  elemental.OperationInfo,
        "path":       req.URL.Path,
    }).Debug("Handling info {{ specification.entity_name|lower }} request.")

    bahamut.WriteHTTPError(w, http.StatusNotImplemented, elemental.NewError("Not implemented", "Info{{ specification.entity_name }} not implemented in Cid yet", "http", http.StatusNotImplemented))
}
