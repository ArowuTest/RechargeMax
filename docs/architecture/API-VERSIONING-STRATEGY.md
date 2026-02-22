# RechargeMax API Versioning Strategy

**Date:** February 1, 2026  
**Current Version:** v1  
**Status:** Active

---

## Overview

This document outlines the API versioning strategy for RechargeMax to ensure smooth transitions when introducing breaking changes while maintaining backward compatibility for existing clients.

---

## Versioning Approach

### URL-Based Versioning

RechargeMax uses **URL-based versioning** where the version is part of the URL path:

```
https://api.rechargemax.ng/api/v1/...
https://api.rechargemax.ng/api/v2/...
```

**Rationale:**
- Clear and explicit
- Easy to implement
- Simple for clients to understand
- Works well with API gateways and proxies

---

## Version Lifecycle

### Stages

1. **Active** - Current production version, fully supported
2. **Deprecated** - Still functional but discouraged, maintenance only
3. **Sunset** - Scheduled for removal, critical fixes only
4. **Retired** - No longer available

### Timeline

```
Active (12+ months) → Deprecated (6 months) → Sunset (3 months) → Retired
```

**Example:**
- **Jan 2026:** v1 Active
- **Jan 2027:** v2 released, v1 becomes Deprecated
- **Jul 2027:** v1 enters Sunset period
- **Oct 2027:** v1 Retired

---

## What Constitutes a Breaking Change

### Breaking Changes (Require New Version)

- Removing an endpoint
- Removing a field from response
- Changing field data type
- Changing authentication method
- Changing error response format
- Renaming fields
- Changing URL structure
- Changing required parameters

### Non-Breaking Changes (Same Version)

- Adding new endpoints
- Adding new optional parameters
- Adding new fields to response
- Adding new error codes
- Improving performance
- Bug fixes
- Documentation updates

---

## Version Support Policy

### Current Version (v1)

**Status:** Active  
**Release Date:** January 2026  
**Support Until:** January 2028 (minimum)

**Features:**
- OTP authentication
- Airtime and data recharge
- Wheel spin game
- Draw entries
- Affiliate program
- Admin dashboard

---

### Future Versions

#### v2 (Planned)

**Tentative Release:** Q1 2027

**Planned Changes:**
- Enhanced authentication (OAuth2 support)
- Improved error handling
- Standardized route naming (RESTful)
- Webhook improvements
- GraphQL endpoint (optional)

**Migration Path:**
- v1 will remain active for 12 months after v2 release
- Migration guide will be provided
- Dual-version support during transition

---

## Deprecation Process

### 1. Announcement (6 months before deprecation)

- Email notification to all API users
- Blog post on developer portal
- Warning headers in API responses:
  ```http
  Warning: 299 - "API v1 will be deprecated on 2027-01-01"
  Sunset: 2027-10-01
  ```

### 2. Deprecation (6 months period)

- API continues to work normally
- Deprecation warnings in responses
- Migration documentation available
- Support team assists with migration

### 3. Sunset (3 months period)

- API still functional but with reduced support
- Only critical security fixes
- Aggressive warnings:
  ```http
  Warning: 299 - "API v1 will be retired on 2027-10-01. Migrate to v2 immediately."
  ```

### 4. Retirement

- API endpoints return 410 Gone
- Response includes migration information:
  ```json
  {
    "error": "API v1 has been retired",
    "code": "VERSION_RETIRED",
    "current_version": "v2",
    "migration_guide": "https://docs.rechargemax.ng/migration/v1-to-v2"
  }
  ```

---

## Version Headers

### Request Headers

Clients can specify version preference:
```http
Accept-Version: v1
```

If not specified, defaults to latest stable version.

### Response Headers

All responses include version information:
```http
API-Version: v1
API-Status: active
API-Deprecation-Date: 2027-01-01
API-Sunset-Date: 2027-10-01
```

---

## Migration Support

### Migration Guides

For each major version transition, we provide:

1. **Migration Guide Document**
   - Breaking changes list
   - Before/after code examples
   - Step-by-step migration instructions

2. **Migration Tools**
   - Request/response comparison tool
   - Automated migration scripts (where possible)
   - Testing sandbox for new version

3. **Support**
   - Dedicated migration support channel
   - Office hours with engineering team
   - Priority support tickets

---

## Backward Compatibility

### Within Same Version

- All changes within a version are backward compatible
- New fields are optional
- Existing fields maintain same behavior
- No breaking changes

### Across Versions

- Multiple versions supported simultaneously
- Clients can migrate at their own pace
- No forced upgrades during active period

---

## Version Detection

### Automatic Detection

API automatically detects version from URL:
```
/api/v1/... → Version 1
/api/v2/... → Version 2
```

### Version Mismatch

If client requests unsupported version:
```http
GET /api/v99/users
```

Response:
```json
{
  "error": "Unsupported API version",
  "code": "UNSUPPORTED_VERSION",
  "supported_versions": ["v1", "v2"],
  "current_version": "v2"
}
```

---

## Communication Channels

### Developer Portal

- **URL:** https://developers.rechargemax.ng
- Version status dashboard
- Migration guides
- API changelog

### Notifications

1. **Email** - All registered API users
2. **Webhook** - Version deprecation events
3. **Blog** - Major announcements
4. **Status Page** - Real-time version status

---

## Monitoring

### Version Usage Metrics

Track for each version:
- Active API keys
- Request volume
- Error rates
- Client distribution

### Deprecation Metrics

Monitor during deprecation:
- Migration progress
- Clients still on old version
- Support ticket volume
- Error patterns

---

## Emergency Procedures

### Critical Security Vulnerability

If a critical security issue is discovered:

1. **Immediate Fix** - Patch all supported versions
2. **Notification** - Email all users within 24 hours
3. **Documentation** - Update security advisories
4. **Accelerated Retirement** - If fix is not possible, accelerate retirement timeline

### Breaking Bug Fix

If a bug fix requires breaking change:

1. **Assess Impact** - Determine affected clients
2. **Communication** - Direct outreach to affected users
3. **Grace Period** - Provide minimum 30 days notice
4. **Fallback** - Offer temporary workaround if possible

---

## Best Practices for Clients

### 1. Always Specify Version

```http
GET /api/v1/users
```

Don't rely on default version.

### 2. Monitor Response Headers

Check for deprecation warnings:
```http
Warning: 299 - "..."
API-Deprecation-Date: ...
```

### 3. Subscribe to Updates

- Join developer mailing list
- Follow changelog
- Monitor status page

### 4. Test New Versions Early

- Use sandbox environment
- Test during deprecation period
- Don't wait until sunset

### 5. Handle Version Errors

```javascript
if (response.code === 'VERSION_RETIRED') {
  // Upgrade to new version
}
```

---

## Governance

### Version Release Authority

- **Engineering Team** - Proposes new versions
- **Product Team** - Approves breaking changes
- **CTO** - Final approval for major versions

### Change Review Process

1. **Proposal** - Document breaking changes
2. **Impact Assessment** - Analyze client impact
3. **Review** - Engineering + Product review
4. **Approval** - CTO approval required
5. **Communication** - Announce to users
6. **Implementation** - Develop and test
7. **Release** - Staged rollout

---

## Exceptions

### Emergency Hotfixes

Critical security or data integrity issues may bypass normal versioning process with:
- Immediate deployment
- Retroactive documentation
- Post-incident review

### Experimental Features

Experimental features may be:
- Released under `/api/v1/experimental/...`
- Subject to change without notice
- Clearly marked as experimental
- Not covered by version guarantees

---

## Contact

For questions about API versioning:
- **Email:** api@rechargemax.ng
- **Slack:** #api-versioning
- **Support:** https://support.rechargemax.ng

---

**Last Updated:** February 1, 2026  
**Next Review:** August 1, 2026
