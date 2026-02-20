#!/bin/bash
echo "================================================"
echo "  RechargeMax JWT Secret Generator"
echo "================================================"
echo ""
echo "Generating cryptographically secure JWT secret..."
echo ""

SECRET=$(openssl rand -hex 32)

echo "✅ Generated 64-character hex secret:"
echo ""
echo "JWT_SECRET=$SECRET"
echo ""
echo "================================================"
echo "  Add this to your .env file"
echo "================================================"
echo ""
echo "Secret length: ${#SECRET} characters"
echo "Entropy: 256 bits"
echo ""
