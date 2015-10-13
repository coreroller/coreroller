import React from "react"
import { Tooltip } from "react-bootstrap"

export const tooltipSafeMode = <Tooltip><strong>Safe mode:</strong> If safe mode is enabled, when a new rollout starts only one instance will be granted an updated, and if it doesn’t succeed updates will be disabled in the group automatically.</Tooltip>
export const tooltipOfficeHours = <Tooltip><strong>Office hours (9am - 5pm):</strong> Use this option to disable updates out of office hours. When using this option, you <strong>must provide a valid timezone</strong> as the office hours time range will be calculated based on it.</Tooltip>
