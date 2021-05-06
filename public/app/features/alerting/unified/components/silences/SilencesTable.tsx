import React, { FC } from 'react';
import { GrafanaTheme2 } from '@grafana/data';
import { Icon, useStyles2, Link, Button, Field } from '@grafana/ui';
import { css } from '@emotion/css';
import { AlertmanagerAlert, Silence } from 'app/plugins/datasource/alertmanager/types';
import SilenceTableRow from './SilenceTableRow';
import { getAlertTableStyles } from '../../styles/table';
import { NoSilencesSplash } from './NoSilencesCTA';
import { AlertManagerPicker } from '../AlertManagerPicker';
import { makeAMLink } from '../../utils/misc';
interface Props {
  silences: Silence[];
  alertManagerAlerts: AlertmanagerAlert[];
  alertManagerSourceName: string;
  setAlertManagerSourceName(name: string): void;
}

const SilencesTable: FC<Props> = ({
  silences,
  alertManagerAlerts,
  alertManagerSourceName,
  setAlertManagerSourceName,
}) => {
  const styles = useStyles2(getStyles);
  const tableStyles = useStyles2(getAlertTableStyles);

  const findSilencedAlerts = (id: string) => {
    return alertManagerAlerts.filter((alert) => alert.status.silencedBy.includes(id));
  };

  return (
    <>
      <Field label="Choose alert manager">
        <AlertManagerPicker current={alertManagerSourceName} onChange={setAlertManagerSourceName} />
      </Field>
      {!!silences.length && (
        <>
          <Link href={makeAMLink('/alerting/silence/new', alertManagerSourceName)}>
            <Button className={styles.addNewSilence} icon="plus">
              New Silence
            </Button>
          </Link>
          <table className={tableStyles.table}>
            <colgroup>
              <col className={tableStyles.colExpand} />
              <col className={styles.colState} />
              <col className={styles.colMatchers} />
              <col />
              <col />
              <col />
            </colgroup>
            <thead>
              <tr>
                <th />
                <th>State</th>
                <th>Matchers</th>
                <th>Alerts</th>
                <th>Schedule</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {silences.map((silence, index) => {
                const silencedAlerts = findSilencedAlerts(silence.id);
                return (
                  <SilenceTableRow
                    key={silence.id}
                    silence={silence}
                    className={index % 2 === 0 ? tableStyles.evenRow : undefined}
                    silencedAlerts={silencedAlerts}
                    alertManagerSourceName={alertManagerSourceName}
                  />
                );
              })}
            </tbody>
          </table>
          <div className={styles.callout}>
            <Icon className={styles.calloutIcon} name="info-circle" />
            <span>Expired silences are automatically deleted after 5 days.</span>
          </div>
        </>
      )}
      {!silences.length && <NoSilencesSplash alertManagerSourceName={alertManagerSourceName} />}
    </>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  addNewSilence: css`
    margin-bottom: ${theme.spacing(1)};
  `,
  colState: css`
    width: 110px;
  `,
  colMatchers: css`
    width: 50%;
  `,
  callout: css`
    background-color: ${theme.colors.background.secondary};
    border-top: 3px solid ${theme.colors.info.border};
    border-radius: 2px;
    height: 62px;
    display: flex;
    flex-direction: row;
    align-items: center;
    margin-top: ${theme.spacing(2)};

    & > * {
      margin-left: ${theme.spacing(1)};
    }
  `,
  calloutIcon: css`
    color: ${theme.colors.info.text};
  `,
});

export default SilencesTable;
