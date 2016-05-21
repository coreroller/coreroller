import API from "../api/API"
import moment from "moment"
import _ from "underscore"
import Store from './BaseStore'

class ActivityStore extends Store {

  constructor() {
    super()
    this.activity = null
    this.getActivity()

    setInterval(() => {
      this.getActivity()
    }, 60 * 1000)
  }

  getCachedActivity() {
    return this.activity
  }

  getActivity() {
    API.getActivity().
      done(activity => {
        this.activity = this.sortActivityByDate(activity)
        this.emitChange()
      }).
      fail((error) => {
        if (error.status === 404) {
          this.activity = []
          this.emitChange()
        }
      })
  }

  sortActivityByDate(entries) {
    let sortedEntries = {}

    entries.forEach(entry => {
      let date = moment.utc(entry.created_ts).local().format("dddd, DD MMMM YYYY")
      if (_.has(sortedEntries, date)) {
        sortedEntries[date].push(entry)
      }
      else {
        sortedEntries[date] = [entry]
      }
    })

    return sortedEntries
  }

  getActivityEntryClass(classID, entry) {

    let classType = {
      1: {
        type: "activityPackageNotFound",
        appName: entry.application_name,
        groupName: entry.group_name,
        channelName: entry.channel_name,
        description: "An update request could not be processed because the group's channel is not linked to any package"
      },
      2: {
        type: "activityRolloutStarted",
        appName: entry.application_name,
        groupName: entry.group_name,
        channelName: entry.channel_name,
        description: "Version " + entry.version + " roll out started"
      },
      3: {
        type: "activityRolloutFinished",
        appName: entry.application_name,
        groupName: entry.group_name,
        channelName: entry.channel_name,
        description: "Version " + entry.version + " successfully rolled out"
      },
      4: {
        type: "activityRolloutFailed",
        appName: entry.application_name,
        groupName: entry.group_name,
        channelName: entry.channel_name,
        description: "There was an error rolling out version " + entry.version + " as the first update attempt failed. Group's updates have been disabled"
      },
      5: {
        type: "activityInstanceUpdateFailed",
        appName: entry.application_name,
        groupName: entry.group_name,
        channelName: entry.channel_name,
        description: "Instance " + entry.instance_id + " reported an error while processing update to version " + entry.version
      },
      6: {
        type: "activityChannelPackageUpdated",
        appName: entry.application_name,
        groupName: entry.group_name,
        channelName: entry.channel_name,
        description: "Channel " + entry.channel_name + " is now pointing to version " + entry.version
      }
    }

    let classDetails = classID ? classType[classID] : classType[1]
    return classDetails
  }

  getActivityEntrySeverity(severityID) {

    let severityType = {
      1: {
        type: "activitySuccess",
        className: "success",
        icon: "fa-check"
      },
      2: {
        type: "activityInfo",
        className: "info",
        icon: "fa-info"
      },
      3: {
        type: "activityWarning",
        className: "warning",
        icon: "fa-exclamation"
      },
      4: {
        type: "activityError",
        className: "error",
        icon: "fa-close"
      }
    }

    let severityInfo = severityID ? severityType[severityID] : severityType[1]
    return severityInfo
  }
}

export default ActivityStore
