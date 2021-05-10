export const RULER_NOT_SUPPORTED_MSG = 'ruler not supported';

export const RULE_LIST_POLL_INTERVAL_MS = 20000;

export const ALERTMANAGER_NAME_QUERY_KEY = 'alertmanager';
export const ALERTMANAGER_NAME_LOCAL_STORAGE_KEY = 'alerting-alertmanager';
export const SILENCES_POLL_INTERVAL_MS = 20000;

export enum Annotation {
  description = 'description',
  summary = 'summary',
  runbookURL = 'runbook_url',
  dashboardUID = '__dashboardUId__',
  panelID = '__panelId__',
}

export const annotationLabels: Record<Annotation, string> = {
  [Annotation.description]: 'Description',
  [Annotation.summary]: 'Summary',
  [Annotation.runbookURL]: 'Runbook URL',
  [Annotation.dashboardUID]: 'Dashboard UID',
  [Annotation.panelID]: 'Panel ID',
};
