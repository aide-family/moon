/* eslint-disable jsx-a11y/iframe-has-title */
import { Space, Grid } from "@arco-design/web-react";
import dayjs from "dayjs";
import React, { useEffect } from "react";

const { Row, Col } = Grid;

const System = () => {
  const [theme, setTheme] = React.useState(
    (localStorage.getItem("theme") as string) || "light"
  );

  const [from, setFrom] = React.useState<number>(
    dayjs().subtract(1, "hour").valueOf()
  );
  const [to, setTo] = React.useState<number>(dayjs().valueOf());

  window.microApp?.addGlobalDataListener(
    (data: { theme?: string; spaceId?: string }) => {
      const { theme } = data;
      if (theme) {
        setTheme(localStorage.getItem("theme") as string);
      }
    }
  );

  useEffect(() => {
    setInterval(() => {
      setFrom(dayjs().subtract(1, "hour").valueOf());
      setTo(dayjs().valueOf());
    }, 1000 * 60 * 5);
  }, [from]);

  return (
    <>
      <Row>
        <Col span={24}>
          <iframe
            src={`https://grafana.bitrecx.com/d-solo/aka/node-exporter-dashboard?orgId=1&refresh=1m&var-vendor=&var-account=&var-group=All&var-name=All&var-instance=124.223.104.203&var-interval=2m&var-total=2&var-device=All&var-maxmount=%2Fetc%2Fhostname&var-show_name=&var-iid=&var-sname=&from=${from}&to=${to}&theme=${theme}&panelId=198`}
            width="100%"
            height="200"
            frameBorder="0"
          ></iframe>
        </Col>
      </Row>
      <Space wrap size={[8, 2]}>
        <iframe
          src={`https://grafana.bitrecx.com/d-solo/aka/node-exporter-dashboard?orgId=1&refresh=1m&var-vendor=&var-account=&var-group=All&var-name=All&var-instance=124.223.104.203&var-interval=2m&var-total=2&var-device=All&var-maxmount=%2Fetc%2Fhostname&var-show_name=&var-iid=&var-sname=&from=${from}&to=${to}&theme=${theme}&panelId=7`}
          width="440"
          height="400"
          frameBorder="0"
        ></iframe>
        <iframe
          src={`https://grafana.bitrecx.com/d-solo/aka/node-exporter-dashboard?orgId=1&refresh=1m&var-vendor=&var-account=&var-group=All&var-name=All&var-instance=124.223.104.203&var-interval=2m&var-total=2&var-device=All&var-maxmount=%2Fetc%2Fhostname&var-show_name=&var-iid=&var-sname=&from=${from}&to=${to}&theme=${theme}&panelId=156`}
          width="440"
          height="400"
          frameBorder="0"
        ></iframe>
        <iframe
          src={`https://grafana.bitrecx.com/d-solo/aka/node-exporter-dashboard?orgId=1&refresh=1m&var-vendor=&var-account=&var-group=All&var-name=All&var-instance=124.223.104.203&var-interval=2m&var-total=2&var-device=All&var-maxmount=%2Fetc%2Fhostname&var-show_name=&var-iid=&var-sname=&from=${from}&to=${to}&theme=${theme}&panelId=157`}
          width="440"
          height="400"
          frameBorder="0"
        ></iframe>
        <iframe
          src={`https://grafana.bitrecx.com/d-solo/aka/node-exporter-dashboard?orgId=1&refresh=1m&var-vendor=&var-account=&var-group=All&var-name=All&var-instance=124.223.104.203&var-interval=2m&var-total=2&var-device=All&var-maxmount=%2Fetc%2Fhostname&var-show_name=&var-iid=&var-sname=&from=${from}&to=${to}&theme=${theme}&panelId=13`}
          width="440"
          height="400"
          frameBorder="0"
        ></iframe>
      </Space>
      <Row>
        <Col span={24}>
          <iframe
            src={`https://grafana.bitrecx.com/d-solo/aka/node-exporter-dashboard?orgId=1&refresh=1m&var-vendor=&var-account=&var-group=All&var-name=All&var-instance=124.223.104.203&var-interval=2m&var-total=2&var-device=All&var-maxmount=%2Fetc%2Fhostname&var-show_name=&var-iid=&var-sname=&from=${from}&to=${to}&theme=${theme}&panelId=158`}
            width="100%"
            height="300"
            frameBorder="0"
          ></iframe>
        </Col>
      </Row>
    </>
  );
};

export default System;
