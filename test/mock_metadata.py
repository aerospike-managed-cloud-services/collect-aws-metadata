"""
AWS metadata HTTP service

Returns only a few specific endpoints (maintenance events and instance-id)
"""

from datetime import datetime, timedelta
import json
import random
import string
import uuid

from fastapi import FastAPI, Header, Response, HTTPException
from fastapi.responses import PlainTextResponse
from pydantic import BaseModel, Field
from typing import Optional


# time window from 10(-ish) days out to 11(-ish) days out; UTC representation
in241h = lambda: (datetime.now() + timedelta(hours=241)).astimezone()
in265h = lambda: (datetime.now() + timedelta(hours=265)).astimezone()

# AWS does not zero-pad the day-of-month but does pad the other numbers (sigh)
# Also, AWS uses "GMT" which does not mean the same thing as UTC. We're
# going to assume they're actually the same time, though, and just print "GMT" at the end.
FMT = "%-d %b %Y %H:%M:%S GMT"

INSTANCE_ID = 'i-0da06b32c373fdecz'

# Store active tokens
active_tokens = set()


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
    NotBefore: str = Field(default_factory=lambda: in241h().strftime(FMT))
    NotAfter: str = Field(default_factory=lambda: in265h().strftime(FMT))
    EventId: str = Field(default_factory=lambda: f"instance-event-{pwd()}")
    State: str = "active"


app = FastAPI()


@app.get("/latest/meta-data/events/maintenance/scheduled", response_class=PlainTextResponse)
async def scheduled(x_aws_ec2_metadata_token: Optional[str] = Header(None)):
    """
    Simulate some scheduled events

    Returns 0-3 events with equal frequency distribution
    """
    if x_aws_ec2_metadata_token and x_aws_ec2_metadata_token not in active_tokens:
        raise HTTPException(status_code=401, detail="Unauthorized")

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


@app.put("/latest/api/token")
async def get_token(response: Response, x_aws_ec2_metadata_token_ttl_seconds: Optional[str] = Header(None)):
    if not x_aws_ec2_metadata_token_ttl_seconds:
        raise HTTPException(status_code=400, detail="Missing TTL header")

    token = str(uuid.uuid4())
    active_tokens.add(token)
    return Response(content=token, media_type="text/plain")


@app.get("/1.0/meta-data/instance-id", response_class=PlainTextResponse)
async def get_instance_id(x_aws_ec2_metadata_token: Optional[str] = Header(None)):
    """
    A hardcoded instance-id for our fake instance
    """
    if x_aws_ec2_metadata_token and x_aws_ec2_metadata_token not in active_tokens:
        raise HTTPException(status_code=401, detail="Unauthorized")
    return INSTANCE_ID
