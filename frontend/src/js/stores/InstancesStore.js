import API from "../api/API"
import Store from "./BaseStore"
import _ from "underscore"

class InstancesStore extends Store {

  constructor() {
    super()
    this.instances = null
    this.getInstanceStatus = this.getInstanceStatus.bind(this)
  }

  getCachedInstances(applicationID, groupID) {
    let cachedInstances = []
    if (this.instances.hasOwnProperty(applicationID)) {
      if (this.instances[applicationID].hasOwnProperty(groupID)) {
        cachedInstances = this.instances[applicationID][groupID]
      }
    }
    return cachedInstances
  }

  getInstances(applicationID, groupID, selectedInstance) {
    let application = this.instances.hasOwnProperty(applicationID) ? this.instances[applicationID] : this.instances[applicationID] = {}

    API.getInstances(applicationID, groupID).
      done(instances => {
        let sortedInstances = _.sortBy(instances, (instance) => {
          instance.statusInfo = this.getInstanceStatus(instance.application.status, instance.application.version)
          return instance.application.last_check_for_updates
        })
        application[groupID] = sortedInstances.reverse()
        this.emitChange()

        if (selectedInstance) {
          this.getInstanceStatusHistory(applicationID, groupID, selectedInstance)
        }
      }).
      fail((error) => {
        if (error.status === 404) {
          application[groupID] = []
          this.emitChange()
        }
      })
  }

  getInstanceStatusHistory(applicationID, groupID, instanceID) {
    let instancesList = this.instances[applicationID][groupID]
    let instanceToUpdate = _.findWhere(instancesList, {id: instanceID})

    return API.getInstanceStatusHistory(applicationID, groupID, instanceID).
      done(statusHistory => {
        instanceToUpdate.statusHistory = statusHistory
        this.emitChange()
      }).
      fail((error) => {
        if (error.status === 404) {
          instanceToUpdate.statusHistory = []
          this.emitChange()
        }
      })
  }

  getInstanceStatus(statusID, version) {
    let status = {
      1: {
        type: "InstanceStatusUndefined",
        className: "",
        spinning: false,
        icon: "",
        description: "",
        status: "Undefined",
        explanation: ""
      },
      2: {
        type: "InstanceStatusUpdateGranted",
        className: "warning",
        spinning: true,
        icon: "",
        description: "Updating: granted",
        status: "Granted",
        explanation: "The instance has received an update package -version " + version + "- and the update process is about to start"
      },
      3: {
        type: "InstanceStatusError",
        className: "danger",
        spinning: false,
        icon: "glyphicon glyphicon-remove",
        description: "Error updating",
        status: "Error",
        explanation: "The instance reported an error while updating to version " + version
      },
      4: {
        type: "InstanceStatusComplete",
        className: "success",
        spinning: false,
        icon: "glyphicon glyphicon-ok",
        description: "Update completed",
        status: "Completed",
        explanation: "The instance has been updated successfully and is now running version " + version
      },
      5: {
        type: "InstanceStatusInstalled",
        className: "warning",
        spinning: true,
        icon: "",
        description: "Updating: installed",
        status: "Installed",
        explanation: "The instance has installed the update package -version " + version + "- but it isn’t using it yet"
      },
      6: {
        type: "InstanceStatusDownloaded",
        className: "warning",
        spinning: true,
        icon: "",
        description: "Updating: downloaded",
        status: "Downloaded",
        explanation: "The instance has downloaded the update package -version " + version + "- and will install it now"
      },
      7: {
        type: "InstanceStatusDownloading",
        className: "warning",
        spinning: true,
        icon: "",
        description: "Updating: downloading",
        status: "Downloading",
        explanation: "The instance has just started downloading the update package -version " + version + "-"
      },
      8: {
        type: "InstanceStatusOnHold",
        className: "default",
        spinning: false,
        icon: "",
        description: "Waiting...",
        status: "On hold",
        explanation: "There was an update pending for the instance but it was put on hold because of the rollout policy"
      }
    }

    let statusDetails = statusID ? status[statusID] : status[1]

    return statusDetails
  }
}

export default InstancesStore
