# GDPR Compliance Guide for Kapok

## Overview

This document outlines Kapok's compliance with the General Data Protection Regulation (GDPR) EU 2016/679.

## Principles of GDPR (Article 5)

### 1. Lawfulness, Fairness and Transparency
- **Legal Basis**: Consent, contract performance, legitimate interests
- **Privacy Notice**: Clear, accessible, written in plain language
- **Transparency**: Data processing activities documented and disclosed

### 2. Purpose Limitation
- Data collected only for specified, explicit, legitimate purposes
- No further processing incompatible with original purpose
- Purpose documented for each data category

### 3. Data Minimization
- Only data necessary for stated purpose is collected
- Regular reviews to identify unnecessary data
- Automated data minimization checks

### 4. Accuracy
- Reasonable steps taken to ensure data accuracy
- Inaccurate data corrected or erased without delay
- Users can update their own data

### 5. Storage Limitation
- Data retained only as long as necessary
- Retention periods defined per data type
- Automated deletion after retention period

### 6. Integrity and Confidentiality
- **Encryption**: AES-256 at rest, TLS 1.3 in transit
- **Access Controls**: RBAC with least privilege
- **Audit Logging**: Immutable audit trail
- **Security Testing**: Regular penetration testing

### 7. Accountability
- Data Protection Officer (DPO) appointed
- Data Protection Impact Assessments (DPIAs) conducted
- Records of processing activities maintained
- Compliance documentation current

## Data Subject Rights

### Right of Access (Article 15)

**Implementation**:
- API endpoint: `GET /api/users/{user_id}/data-export`
- Response format: JSON with all personal data
- Response time: Within 30 days (legally required)
- Actual target: Within 24 hours

**Data Included**:
- User profile information
- Tenant associations
- Audit log entries
- Historical data changes

### Right to Rectification (Article 16)

**Implementation**:
- API endpoint: `PUT /api/users/{user_id}`
- Users can update their own data
- Admins can correct inaccuracies
- Changes logged in audit trail

### Right to Erasure / "Right to be Forgotten" (Article 17)

**Implementation**:
- API endpoint: `DELETE /api/users/{user_id}/gdpr-delete`
- Soft delete: User data anonymized
- Hard delete: Data cryptographically erased
- Deletion confirmed within 30 days
- Exceptions: Legal obligations, audit logs retained

**Deletion Process**:
1. User submits deletion request
2. Verification of identity
3. Legal review (check for retention obligations)
4. Data anonymization or erasure
5. Confirmation sent to user

### Right to Restriction of Processing (Article 18)

**Implementation**:
- User can request processing restriction
- Account marked as "restricted"
- Data retained but not processed
- Restrictions lifted upon user request

### Right to Data Portability (Article 20)

**Implementation**:
- API endpoint: `GET /api/users/{user_id}/data-export?format=json`
- Structured, machine-readable format (JSON)
- Includes all user-provided data
- Can be imported into another system

### Right to Object (Article 21)

**Implementation**:
- Object to processing for direct marketing
- Object to automated decision-making
- Objection form available in settings
- Processing stopped within 1 month

### Rights Related to Automated Decision Making (Article 22)

**Implementation**:
- No fully automated decisions with legal effects
- All automated processes have human oversight
- Users informed of automated processing
- Right to request human review

## Lawful Basis for Processing

### Consent (Article 6(1)(a))
- **Explicit Consent**: Checkbox, not pre-ticked
- **Granular**: Separate consent for each purpose
- **Withdrawal**: Easy to withdraw consent
- **Records**: Consent timestamp and IP logged

### Contract Performance (Article 6(1)(b))
- Processing necessary to provide service
- User creates account = contractual relationship
- Data processing essential for service delivery

### Legal Obligation (Article 6(1)(c))
- Tax records: 7 years retention
- Audit logs: 7 years retention (HIPAA/SOC2)
- Financial transaction records

### Legitimate Interests (Article 6(1)(f))
- Fraud prevention
- Network and information security
- Internal administration

## Special Categories of Personal Data (Article 9)

### Health Data (if applicable)
- **Explicit Consent** or **Healthcare provision** basis
- Enhanced security controls
- Access restricted to authorized personnel
- Logged and monitored

### Other Special Categories
- Racial/ethnic origin: Not collected
- Political opinions: Not collected
- Religious beliefs: Not collected
- Trade union membership: Not collected
- Genetic/biometric data: Only if WebAuthn used (fingerprint)
- Sexual orientation: Not collected

## Data Protection by Design and Default (Article 25)

### Design Measures
- **Pseudonymization**: User IDs instead of names in logs
- **Encryption**: Default for all data
- **Minimization**: Collect only necessary fields
- **Access Controls**: RBAC by default

### Default Measures
- Privacy settings default to most restrictive
- Optional data fields not required
- Cookie consent required (no pre-ticked boxes)
- Data retention limits enforced

## Data Protection Impact Assessment (DPIA)

### When Required
- Large-scale processing of special category data
- Systematic monitoring of public areas
- Automated decision-making with legal effects
- High risk to rights and freedoms

### DPIA Process
1. **Describe Processing**: Purpose, data, recipients, retention
2. **Assess Necessity**: Legitimate interests, necessity check
3. **Identify Risks**: Privacy risks to data subjects
4. **Mitigation Measures**: Controls to reduce risks
5. **DPO Review**: Consultation with Data Protection Officer
6. **Documentation**: Record and maintain DPIA

## Data Breach Notification (Article 33-34)

### Internal Breach Response
1. **Detection**: Automated monitoring, user reports
2. **Assessment**: Severity, scope, affected data subjects
3. **Containment**: Isolate, prevent further breach
4. **Investigation**: Root cause analysis

### Supervisory Authority Notification (Article 33)
- **Timeframe**: Within 72 hours of becoming aware
- **Contents**: Nature of breach, affected data, consequences, mitigation
- **Authority**: Data Protection Authority in relevant EU member state

### Data Subject Notification (Article 34)
- **When Required**: High risk to rights and freedoms
- **Timeframe**: Without undue delay
- **Contents**: Nature of breach, DPO contact, likely consequences, mitigation
- **Method**: Direct communication (email, in-app notification)

### Breach Documentation
- Record of all breaches (even if not reportable)
- Facts of breach, effects, remedial action
- Maintained for supervisory authority review

## International Data Transfers (Article 44-50)

### Transfer Mechanisms

**EU to Third Countries**
- **Adequacy Decision**: EU Commission approved countries
- **Standard Contractual Clauses (SCCs)**: With non-EU processors
- **Binding Corporate Rules (BCRs)**: For multinational organizations

### Data Residency
- **EU Data**: Stored in EU regions (Ireland, Frankfurt, Paris)
- **User Choice**: Select geographic region for data storage
- **No Cross-Border Transfers**: Without appropriate safeguards

## Records of Processing Activities (Article 30)

### Data Controller Information
- Name and contact: Kapok Platform
- DPO contact: dpo@kapok.io
- Purposes of processing
- Categories of data subjects and personal data
- Recipients of personal data
- International transfers (if any)
- Retention periods
- Security measures

### Data Processor Information
- Cloud providers (AWS/GCP/Azure)
- Monitoring services
- Email service providers
- Payment processors (if applicable)

## Data Protection Officer (DPO)

### DPO Responsibilities
- Monitor GDPR compliance
- Advise on Data Protection Impact Assessments
- Cooperate with supervisory authorities
- Act as point of contact for data subjects

### DPO Contact
**Email**: dpo@kapok.io
**Response Time**: Within 5 business days

## Third-Party Processors

### Processor Requirements
- Data Processing Agreement (DPA) signed
- GDPR compliance verified
- Security assessment conducted
- Sub-processor approval required

### Current Processors
1. **Cloud Infrastructure**: AWS/GCP/Azure (DPA signed)
2. **Monitoring**: Self-hosted (no third-party)
3. **Email**: [To be configured with DPA]

## Employee Training

### Privacy Training
- Annual GDPR training for all employees
- Specialized training for those handling personal data
- Incident response training
- Regular updates on GDPR developments

### Documentation
- Training records maintained
- Attendance tracked
- Test scores recorded
- Re-training for failures

## GDPR Compliance Checklist

- [x] Privacy notice published
- [x] Legal basis documented
- [x] Data minimization implemented
- [x] Consent management system
- [x] Data subject rights endpoints
- [x] Data export functionality
- [x] Right to erasure implemented
- [x] Encryption at rest and in transit
- [x] Access controls (RBAC)
- [x] Audit logging (immutable)
- [x] Breach notification procedures
- [x] DPO appointed
- [x] DPIA process defined
- [x] Data retention policies
- [x] International transfer safeguards
- [x] Records of processing activities
- [x] Employee training program

## Contact Information

**Data Protection Officer**
Email: dpo@kapok.io
Phone: [To be configured]

**Privacy Inquiries**
Email: privacy@kapok.io

**Data Breach Reporting**
Email: security@kapok.io (24/7 monitoring)

## Document Control

- **Version**: 1.0
- **Last Updated**: 2026-01-29
- **Next Review**: 2026-04-29 (Quarterly)
- **Owner**: Data Protection Officer
- **Classification**: Public
