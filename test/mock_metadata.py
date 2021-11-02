"""
AWS metadata HTTP service

Returns only a few specific endpoints (maintenance events and instance-id)
"""

from datetime import datetime, timedelta
import json
import random
import string

from fastapi import FastAPI
from fastapi.responses import PlainTextResponse
from pydantic import BaseModel, Field


# time window from 10 days out to 11 days out; UTC representation
in240h = lambda: (datetime.now() + timedelta(hours=240)).astimezone()
in264h = lambda: (datetime.now() + timedelta(hours=264)).astimezone()

# AWS does not zero-pad the day-of-month but does pad the other numbers (sigh)
# Also, AWS uses "GMT" which does not mean the same thing as UTC. We're
# going to assume they're actually the same time, though, and just print "GMT" at the end.
FMT = "%-d %b %Y %H:%M:%S GMT"

INSTANCE_ID = 'i-0da06b32c373fdecz'


def pwd(n=10):
    """
    A random password of letters + digits
    """
    return ''.join(random.sample(string.ascii_lowercase+string.digits, n))


class Event(BaseModel):
    """
    AWS instance metadata maintenance event structure
    """
    Code: str = "system-reboot"
    Description: str = "scheduled reboot"
    NotBefore: str = Field(default_factory=lambda: in240h().strftime(FMT))
    NotAfter: str = Field(default_factory=lambda: in264h().strftime(FMT))
    EventId: str = Field(default_factory=lambda: f"instance-event-{pwd()}")
    State: str = "active"


app = FastAPI()


@app.get("/latest/meta-data/events/maintenance/scheduled", response_class=PlainTextResponse)
async def scheduled():
    """
    Simulate some scheduled events

    Returns 0-3 events with equal frequency distribution
    """
    count = random.choice(range(4))

    # simple implementation would just return instances of Event() and
    # fastapi would json-encode them for us.
    #
    # but fastapi always adds content-type:application/json and AWS (incorrectly)
    # sets text/plain, which we're trying to emulate. so we're forced to do this
    # thing instead, to allow the emulated header to match AWS.
    ret = json.dumps(
        [Event().dict() for n in range(count)],
        indent=2,
    )

    return ret


@app.get("/1.0/meta-data/instance-id", response_class=PlainTextResponse)
async def instance_id():
    """
    A hardcoded instance-id for our fake instance
    """
    return INSTANCE_ID
