/**
  @typedef {Object} MaintenanceEvent
  @property {Date} maintenance_time_start
  @property {Date} maintenance_time_end
  @property {Date} server_down_start
  @property {Date} server_down_end
  @property {string} description
 */

/**
  @typedef {Object} CustomWindowObject
  @property {any} Chart
  @property {any} Alpine

  @typedef {Window & typeof globalThis & CustomWindowObject} CustomWindow
 */
