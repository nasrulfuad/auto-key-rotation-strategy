#!/usr/bin/env bash
set -euo pipefail

duration="${1:-300}"
my_app="${MY_APP_URL:-http://localhost:8081}"
partner_app="${PARTNER_APP_URL:-http://localhost:8082}"

show_keys() {
  echo "--- $1 signing keys"
  curl -s "$2/signing-keys" | python3 -c 'import json,sys
body=json.load(sys.stdin)
for k in body["data"]:
    print("  %-16.16s  %-8s  active %s-%s  delete %s" % (k["key_id"], k["status"], k["active_from"][11:19], k["active_until"][11:19], k["delete_after"][11:19]))'
}

end=$((SECONDS + duration))
request=0
previous_my_key=""
previous_partner_key=""
while [ "$SECONDS" -lt "$end" ]; do
  request=$((request + 1))
  entitlement="ENT$((request % 3))"
  if [ $((request % 4)) -eq 0 ]; then
    response=$(curl -s -w '\n%{http_code}' -X POST "$my_app/terminations" \
      -H 'Content-Type: application/json' -d "{\"subject_id\":\"SUB001\",\"entitlement_id\":\"$entitlement\"}")
  else
    response=$(curl -s -w '\n%{http_code}' -X POST "$my_app/activations" \
      -H 'Content-Type: application/json' -d "{\"subject_id\":\"SUB001\",\"entitlement_id\":\"$entitlement\",\"products\":\"com.partner.product1\",\"offer\":{\"free_trial\":\"NONE\"}}")
  fi
  status=$(tail -n1 <<<"$response")
  my_key=$(head -n1 <<<"$response" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("signed_with_key_id","-"))' 2>/dev/null || echo "-")
  partner_key=$(curl -s "$my_app/activation-notifications" | python3 -c 'import json,sys; d=json.load(sys.stdin)["data"]; print(d[0]["key_id"] if d else "-")' 2>/dev/null || echo "-")
  printf '%s  #%-4d status=%s  my_app kid=%.16s  partner_app kid=%.16s\n' "$(date +%T)" "$request" "$status" "$my_key" "$partner_key"
  if [ "$my_key" != "$previous_my_key" ] || [ "$partner_key" != "$previous_partner_key" ]; then
    show_keys my_app "$my_app"
    show_keys partner_app "$partner_app"
    previous_my_key="$my_key"
    previous_partner_key="$partner_key"
  fi
  sleep 3
done
