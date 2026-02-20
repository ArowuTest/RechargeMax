#!/bin/bash
cd /home/ubuntu/RechargeMax_Clean/backend
export $(grep -v '^#' .env | xargs)
./rechargemax
