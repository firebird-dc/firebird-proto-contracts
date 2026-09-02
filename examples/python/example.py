#!/usr/bin/env python3
import argparse
import sys
from urllib.parse import urlparse

import grpc
from google.protobuf.json_format import MessageToDict

import user_management_pb2
import user_management_pb2_grpc


def parse_target(api_url: str, insecure: bool) -> tuple[str, bool]:
    """
    Accepts:
      - https://api.example.com
      - http://localhost:9090
      - api.example.com:443
      - localhost:9090

    Returns:
      (grpc_target, use_tls)
    """
    if "://" in api_url:
        parsed = urlparse(api_url)
        host = parsed.hostname
        if not host:
            raise ValueError(f"Invalid api-url: {api_url}")
        port = parsed.port
        if port is None:
            port = 443 if parsed.scheme == "https" else 80
        use_tls = parsed.scheme == "https" and not insecure
        return f"{host}:{port}", use_tls

    # No scheme provided, assume raw gRPC host:port
    return api_url, not insecure


def build_channel(target: str, use_tls: bool) -> grpc.Channel:
    if use_tls:
        return grpc.secure_channel(target, grpc.ssl_channel_credentials())
    return grpc.insecure_channel(target)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Firebird IAM gRPC example: Token -> GetUserInfo"
    )
    parser.add_argument(
        "--api-url",
        required=True,
        help="gRPC host/port or URL, e.g. https://api.example.com or localhost:9090",
    )
    parser.add_argument(
        "--client-id",
        required=True,
        help="Service account client_id for client_credentials grant",
    )
    parser.add_argument(
        "--client-secret",
        required=True,
        help="Service account client_secret for client_credentials grant",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=15.0,
        help="Per-request timeout in seconds",
    )
    parser.add_argument(
        "--insecure",
        action="store_true",
        help="Use insecure gRPC channel (useful for local dev)",
    )
    args = parser.parse_args()

    try:
        target, use_tls = parse_target(args.api_url, args.insecure)
    except ValueError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 2

    print(f"Connecting to gRPC target: {target} (tls={use_tls})")

    channel = build_channel(target, use_tls)
    stub = user_management_pb2_grpc.UserManagementServiceStub(channel)

    try:
        grpc.channel_ready_future(channel).result(timeout=args.timeout)
    except grpc.FutureTimeoutError:
        print("ERROR: gRPC channel was not ready in time.", file=sys.stderr)
        return 3

    # 1) Token
    token_req = user_management_pb2.TokenRequest(
        grant_type="client_credentials",
        client_id=args.client_id,
        client_secret=args.client_secret,
    )

    try:
        token_resp = stub.Token(token_req, timeout=args.timeout)
    except grpc.RpcError as e:
        print(f"Token RPC failed: code={e.code()} details={e.details()}", file=sys.stderr)
        return 4

    access_token = token_resp.access_token
    if not access_token:
        print("ERROR: Token response did not contain access_token.", file=sys.stderr)
        return 5

    print("Token acquired successfully.")
    print(f"token_type={token_resp.token_type}")
    print(f"expires_in={token_resp.expires_in}")
    print(f"access_token_prefix={access_token[:24]}...")

    # 2) GetUserInfo with bearer token
    metadata = (("authorization", f"Bearer {access_token}"),)

    try:
        me_resp = stub.GetUserInfo(
            user_management_pb2.GetUserInfoRequest(),
            timeout=args.timeout,
            metadata=metadata,
        )
    except grpc.RpcError as e:
        print(
            f"GetUserInfo RPC failed: code={e.code()} details={e.details()}",
            file=sys.stderr,
        )
        return 6

    print("\nCurrent user info:")
    print(MessageToDict(me_resp, preserving_proto_field_name=True))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())