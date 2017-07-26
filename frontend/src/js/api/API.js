import $ from "jquery";
import _ from "underscore";
import moment from "moment";
import PubSub from "pubsub-js"

const MAIN_PROGRESS_BAR = "main_progress_bar"
const BASE_URL = "/api"

class API {

  static logout() {
    $.ajax({
      type: "GET",
      url: BASE_URL + "/activity",
      async: false,
      username: "admin",
      password: "invalid-password",
      headers: { "Authorization": "Basic xxx" }
    })
    .fail(function(){
      window.location = "/"
    })
  }

  // Applications

  static getApplications() {  
    return API.getJSON(BASE_URL + "/apps")
  }

  static getApplication(applicationID) {
    return API.getJSON(BASE_URL + "/apps/" + applicationID)
  }

  static deleteApplication(applicationID) {
    let url = BASE_URL + "/apps/" + applicationID
    
    return API.doRequest("DELETE", url, "")
  }

  static createApplication(applicationData, clonedFromAppID) {
    let url = BASE_URL + "/apps"
    if (clonedFromAppID) {
      url += "?clone_from=" + clonedFromAppID
    }

    return API.doRequest("POST", url, JSON.stringify(applicationData))
  }

  static updateApplication(applicationData) {
    let url = BASE_URL + "/apps/" + applicationData.id
    
    return API.doRequest("PUT", url, JSON.stringify(applicationData))
  }

  // Groups

  static getGroup(applicationID, groupID) {
    return API.getJSON(BASE_URL + "/apps/" + applicationID + "/groups/" + groupID)
  }

  static deleteGroup(applicationID, groupID) {
    let url = BASE_URL + "/apps/" + applicationID + "/groups/" + groupID
    
    return API.doRequest("DELETE", url, "")
  }

  static createGroup(groupData) {
    let applicationID = groupData.appID,
        url = BASE_URL + "/apps/" + groupData.application_id + "/groups"
    
    return API.doRequest("POST", url, JSON.stringify(groupData))
  }

  static updateGroup(groupData) {
    let keysToRemove = ["id", "created_ts", "version_breakdown", "instances_stats", "channel"],
        processedGroup = API.removeKeysFromObject(groupData, keysToRemove),
        url = BASE_URL + "/apps/" + groupData.application_id + "/groups/" + groupData.id
    
    return API.doRequest("PUT", url, JSON.stringify(groupData))
  }

  // Channels

  static deleteChannel(applicationID, channelID) {
    let url = BASE_URL + "/apps/" + applicationID + "/channels/" + channelID

    return API.doRequest("DELETE", url, "")
  }

  static createChannel(channelData) {
    let url = BASE_URL + "/apps/" + channelData.application_id + "/channels"

    return API.doRequest("POST", url, JSON.stringify(channelData))
  }

  static updateChannel(channelData, onSuccess) {
    let keysToRemove = ["id", "created_ts", "package"],
        processedChannel = API.removeKeysFromObject(channelData, keysToRemove),
        url = BASE_URL + "/apps/" + channelData.application_id + "/channels/" + channelData.id

    return API.doRequest("PUT", url, JSON.stringify(processedChannel))
  }

  // Packages

  static deletePackage(applicationID, packageID) {
    let url = BASE_URL + "/apps/" + applicationID + "/packages/" + packageID

    return API.doRequest("DELETE", url, "")
  }

  static createPackage(packageData) {
    let url = BASE_URL + "/apps/" + packageData.application_id + "/packages"

    return API.doRequest("POST", url, JSON.stringify(packageData))
  }

  static updatePackage(packageData) {
    let keysToRemove = ["id", "created_ts", "package"],
        processedPackage = API.removeKeysFromObject(packageData, keysToRemove),
        url = BASE_URL + "/apps/" + packageData.application_id + "/packages/" + packageData.id

    return API.doRequest("PUT", url, JSON.stringify(processedPackage))
  }

  // Instances

  static getInstances(applicationID, groupID) {
    let url = BASE_URL + "/apps/" + applicationID + "/groups/" + groupID + "/instances"
    
    return API.getJSON(url)
  }

  static getInstanceStatusHistory(applicationID, groupID, instanceID) {
    let url = BASE_URL + "/apps/" + applicationID + "/groups/" + groupID + "/instances/" + instanceID + "/status_history"
    
    return API.getJSON(url)
  }

  // Activity

  static getActivity() {
    let now = moment().utc().toISOString(),
        weekAgo = moment().utc().subtract(7, "days").toISOString(),
        query = "?start=" + weekAgo + "&end=" + now,
        url = BASE_URL + "/activity" + query

    return API.getJSON(url)
  }

  // User

  static updateUserPassword(userData) {
    let url = BASE_URL + "/password"

    return API.doRequest("PUT", url, JSON.stringify(userData))
  }

  // Helpers

  static removeKeysFromObject(data, valuesToRemove) {
    return _.omit(data, valuesToRemove)
  }

  // Wrappers

  static getJSON(url) {
    PubSub.publish(MAIN_PROGRESS_BAR, "add")

    return $.getJSON(url).
      always(() => { PubSub.publish(MAIN_PROGRESS_BAR, "done") })
  }

  static doRequest(method, url, data) {
    PubSub.publish(MAIN_PROGRESS_BAR, "add")

    return $.ajax({
        method: method,
        url: url,
        data: data,
        dataType: "json"
      }).
      always(() => { PubSub.publish(MAIN_PROGRESS_BAR, "done") })
  }

}

export default API
