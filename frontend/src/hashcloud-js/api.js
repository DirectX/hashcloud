'use strict'

var HashCloudAPI = function (options) {
  if (!(this instanceof HashCloudAPI)) {
    return new HashCloudAPI(options)
  }

  if (options) {
    if (options.apiUrl !== undefined) {
      this.apiUrl = options.apiUrl
    }
  }

  this.test = function () {
    console.log("test", this.apiUrl);
  }
}

module.exports = { HashCloudAPI }