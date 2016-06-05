# RoccaForte

Event Aggregation And Notification.

## Big Picture

```
                        --------------------
 Event ---> (HTTP) ---> |                  |
 Event ---> (Queue) --> |    RoccaForte    |----> Aggregated notification
 Event ---> (Other) --> |                  |
                        --------------------
```

## Aggregation

Events that are of the same type are aggregated by the time received, so that
only one notification in, for example, every 5 minutes, no matter how many
events of the same type are received during that time frame.

## Notification

Notifications bundle multiple events, and carry only enough information for
the clients to query back to RoccaForte what exactly the events were.

## Destination

Destination is anywhere we can deliver our notifications to. The most basic
type of the destination is an HTTP endpoint, but could be anything, really.