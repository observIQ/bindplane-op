import { gql, OnSubscriptionDataOptions } from "@apollo/client";
import {
  Card,
  Chip,
  Collapse,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableRow,
  Typography,
} from "@mui/material";
import { memo, useEffect, useRef, useState } from "react";
import {
  LiveTailRecordType,
  LivetailSubscription,
  useLivetailSubscription,
} from "../../graphql/generated";
import { ChevronDown } from "../Icons";
import styles from "./live-tail-console.module.scss";
import { LTSearchBar } from "./SearchBar";
import { LogRecord, MetricRecord } from "./types";

interface Props {
  ids: string[];
}

type Message = LivetailSubscription["livetail"][0];

gql`
  subscription livetail($ids: [String!]!, $filters: [String!]!) {
    livetail(agentIds: $ids, filters: $filters) {
      type
      records
    }
  }
`;

export const LiveTailConsole: React.FC<Props> = ({ ids }) => {
  const [filters, setFilters] = useState<string[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);

  const consoleRef = useRef<HTMLDivElement | null>(null);
  const lastRowRef = useRef<HTMLDivElement | null>(null);

  const { loading, error } = useLivetailSubscription({
    variables: { ids, filters },
    onSubscriptionData: ({
      subscriptionData,
    }: OnSubscriptionDataOptions<LivetailSubscription>) => {
      const { data } = subscriptionData;
      if (data == null) return;

      console.log({ data });

      const newMessages = [...messages, ...data.livetail];
      const last100 = newMessages.slice(newMessages.length - 100);

      setMessages(last100);
    },
  });

  useEffect(() => {
    if (consoleRef.current) {
      consoleRef.current.scrollTop = consoleRef.current.scrollHeight;
    }
  }, [messages]);

  function handleFilterChange(v: string[]) {
    setFilters(v);
  }

  return (
    <div className={styles.container}>
      <div className={styles.console} ref={consoleRef} onScroll={() => {}}>
        <div className={styles.header}>
          <div className={styles.ch} />
          <div className={styles.dt}>Time</div>
          <div className={styles.lg}>Message</div>
        </div>
        {messages.map((row) => (
          <LiveTailRow message={row} />
        ))}
      </div>
      <div ref={lastRowRef} />

      <LTSearchBar value={filters} onValueChange={handleFilterChange} />
    </div>
  );
};

const LiveTailRowComponent: React.FC<{ message: Message }> = ({ message }) => {
  const { timestamp, ...rest } = message.records[0];
  const [open, setOpen] = useState(false);

  function renderSummary(message: Message) {
    switch (message.type) {
      case LiveTailRecordType.Logs:
        const logRecord = message.records[0] as LogRecord;
        return (
          <Stack direction="row" spacing={2} alignItems="center">
            <Chip label={"info"} size={"small"} />
            <Typography fontFamily="monospace" fontSize={12}>
              {logRecord.body}
            </Typography>
          </Stack>
        );
      case LiveTailRecordType.Metrics:
        const metricRecord = message.records[0] as MetricRecord;
        return (
          <Stack direction="row" spacing={1}>
            <Chip label={metricRecord.name} size={"small"} />
            <Typography fontFamily="monospace" fontSize={12}>
              {metricRecord.value} {metricRecord.unit}
            </Typography>
          </Stack>
        );
      case LiveTailRecordType.Traces:
        return <>trace todo</>;
    }
  }

  return (
    <Card
      key={`${timestamp}-${message.type}`}
      onClick={() => setOpen((prev) => !prev)}
      classes={{ root: styles.card }}
    >
      <Stack direction="row" alignItems="center">
        <div className={styles.ch}>
          <ChevronDown className={styles.chevron} />
        </div>
        <div className={styles.dt}>
          <Typography fontFamily="monospace" fontSize={12} marginLeft={0.2}>
            {formatLogDate(new Date(message.records[0].timestamp))}
          </Typography>
        </div>

        <div className={styles.summary}>{renderSummary(message)}</div>
      </Stack>
      <Collapse in={open}>
        <div className={styles["table-container"]}>
          <Table padding="none" size="small">
            <TableBody>
              {Object.entries(rest).map(([k, v]) => (
                <TableRow>
                  <TableCell className={styles.key}>{k}</TableCell>
                  <TableCell className={styles.value}>
                    <pre>{JSON.stringify(v, undefined, 4)}</pre>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Collapse>
    </Card>
  );
};

const logDateFormat = new Intl.DateTimeFormat(undefined, {
  month: "short",
  day: "2-digit",
  year: undefined,
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  timeZoneName: "short",
});

export function formatLogDate(date: Date): string {
  return logDateFormat.format(date);
}
const LiveTailRow = memo(LiveTailRowComponent);
