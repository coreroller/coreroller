import API from "../api/API"
import Store from './BaseStore'
import _ from "underscore"

class InstancesStore extends Store {

  constructor() {
    super()
    this.instances = {}
  }

  getAll() {
    return this.instances
  }

  getInstances(applicationID, groupID) {
    API.getInstances(applicationID, groupID).
      done(instances => {
        let application = this.instances.hasOwnProperty(applicationID) ? this.instances[applicationID] : this.instances[applicationID] = {}
        application[groupID] = instances
        this.emitChange()
      })
  }

  getInstanceStatusHistory(applicationID, groupID, instanceID) {
    API.getInstanceStatusHistory(applicationID, groupID, instanceID).
      done(statusHistory => {
        console.log(statusHistory)
        let instancesList = this.instances[applicationID][groupID]
        let instanceToUpdate = _.findWhere(instancesList, {id: instanceID})
        instanceToUpdate.statusHistory = statusHistory
        this.emitChange()
      })
  }

  getInstanceStatus(statusID) {
    let status = {
      1: {
        type: "InstanceStatusUndefined",
        className: "",
        spinning: false,
        description: ""
      },
      2: {
        type: "InstanceStatusUpdateGranted",
        className: "warning",
        spinning: true,
        description: "Updating: granted"
      },
      3: {
        type: "InstanceStatusError",
        className: "danger",
        spinning: false,
        description: "Error updating"
      },
      4: {
        type: "InstanceStatusComplete",
        className: "success",
        spinning: false,
        description: "Update completed"
      },
      5: {
        type: "InstanceStatusInstalled",
        className: "warning",
        spinning: true,
        description: "Updating: installed"
      },
      6: {
        type: "InstanceStatusDownloaded",
        className: "warning",
        spinning: true,
        description: "Updating: downloaded"
      },  
      7: {
        type: "InstanceStatusDownloading",
        className: "warning",
        spinning: true,
        description: "Updating: downloading"
      },
      8: {
        type: "InstanceStatusOnHold",
        className: "default",
        spinning: false,
        description: "Waiting..."
      }
    }

    let statusDetails = statusID ? status[statusID] : status[1]

    return statusDetails
  }

}

export default InstancesStore