{{template "base/base.html" .}}
{{define "body"}}
  <div class="right-content-container">
    <div class="header">
      <ol class="breadcrumb">
        <li><a href="/">Home</a></li>
        <li><a href="/schedule">Schedule</a></li>
      </ol>
    </div>

    <div class = "content-block-empty">
      <div class="col-md-12">
        <div class="col-md-12 white-bg box">
          <div class="row">
            <h4>Automate Image Update</h4>
          </div>
        </div>
      </div>
      <div class="col-md-12">
        <div class="col-md-12 white-bg box">
          <div class="row">
            <h4>Automate Image Cleanup</h4>
            <div class="form-group">
              <label for="registry">Select list:</label>
              <select class="form-control" id="registry">
                <option>3</option>
                <option>4</option>
              </select>
            </div>
          </div>
        </div>
      </div>
  </div>

{{end}}
