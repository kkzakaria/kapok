# HIPAA Compliance Guide for Kapok

## Overview

This document outlines Kapok's compliance with the Health Insurance Portability and Accountability Act (HIPAA) for handling Protected Health Information (PHI).

**Important**: Kapok can be configured to be HIPAA-compliant, but requires:
1. Business Associate Agreement (BAA) with customers
2. Proper configuration and usage policies
3. Employee training on HIPAA requirements

## Protected Health Information (PHI)

### What is PHI?

PHI includes any information about health status, provision of healthcare, or payment for healthcare that can be linked to a specific individual.

**18 HIPAA Identifiers**:
1. Names
2. Geographic subdivisions smaller than state
3. Dates (birth, admission, discharge, death)
4. Telephone numbers
5. Fax numbers
6. Email addresses
7. Social Security numbers
8. Medical record numbers
9. Health plan beneficiary numbers
10. Account numbers
11. Certificate/license numbers
12. Vehicle identifiers
13. Device identifiers
14. URLs
15. IP addresses
16. Biometric identifiers
17. Full-face photos
18. Other unique identifying numbers or codes

## HIPAA Security Rule

### Administrative Safeguards (§164.308)

#### Security Management Process (§164.308(a)(1))

**Risk Analysis**
- Annual HIPAA risk assessments conducted
- Vulnerabilities identified and documented
- Risk mitigation plans developed
- Assessment documented and retained

**Risk Management**
- Security controls implemented based on risk
- Continuous monitoring and improvement
- Regular security updates and patches
- Vulnerability remediation tracking

**Sanction Policy**
- Clear consequences for HIPAA violations
- Disciplinary actions documented
- Employee termination for severe violations
- Sanctions applied consistently

**Information System Activity Review**
- Audit logs reviewed monthly
- Automated anomaly detection
- Security incident investigation
- Review findings documented

#### Assigned Security Responsibility (§164.308(a)(2))

**Security Officer**
- Dedicated HIPAA Security Officer appointed
- Responsible for security program oversight
- Reports directly to executive leadership
- Contact: security@kapok.io

#### Workforce Security (§164.308(a)(3))

**Authorization and Supervision**
- Job roles defined with PHI access requirements
- Supervisory oversight of PHI access
- Periodic access reviews
- Least privilege principle enforced

**Workforce Clearance**
- Background checks for PHI access
- Security screening procedures
- Clearance documented and maintained

**Termination Procedures**
- Access revoked immediately upon termination
- Return of security credentials
- Exit interview including HIPAA obligations
- Termination checklist completed

#### Information Access Management (§164.308(a)(4))

**Isolating Healthcare Clearinghouse Functions**
- Not applicable (Kapok is not a clearinghouse)

**Access Authorization**
- Role-Based Access Control (RBAC)
- Approval workflow for PHI access
- Access requests documented
- Periodic access certification

**Access Establishment and Modification**
- Formal provisioning process
- Access changes require approval
- Automated deprovisioning
- Changes logged in audit trail

#### Security Awareness and Training (§164.308(a)(5))

**Security Reminders**
- Quarterly security awareness emails
- Phishing simulation training
- Security bulletins for emerging threats

**Protection from Malicious Software**
- Container image scanning
- Dependency vulnerability scanning
- Regular security updates

**Log-in Monitoring**
- Failed login attempts tracked
- Account lockout after 5 attempts
- Suspicious activity alerts
- Login monitoring dashboard

**Password Management**
- Password policy: 12+ characters, complexity
- Password rotation every 90 days
- Password history: last 12 passwords
- Multi-factor authentication required

#### Security Incident Procedures (§164.308(a)(6))

**Response and Reporting**
1. Detection: Automated monitoring, user reports
2. Assessment: Severity, PHI involved, breach determination
3. Containment: Isolate affected systems
4. Investigation: Root cause analysis
5. Notification: Breach notification if required (60 days)
6. Remediation: Fix vulnerabilities
7. Documentation: Incident report retained for 6 years

#### Contingency Plan (§164.308(a)(7))

**Data Backup Plan**
- Automated daily full backups
- Hourly incremental backups
- Encrypted backup storage
- Off-site backup retention
- Backup testing quarterly

**Disaster Recovery Plan**
- Recovery Time Objective (RTO): 4 hours (Standard), 1 hour (Professional), 15 minutes (Enterprise)
- Recovery Point Objective (RPO): 6 hours (Standard), 1 hour (Professional), 5 minutes (Enterprise)
- Documented recovery procedures
- Annual disaster recovery drill

**Emergency Mode Operation Plan**
- Critical functions identified
- Manual procedures documented
- Emergency contacts maintained
- Communication plan defined

**Testing and Revision**
- Quarterly backup restore testing
- Annual disaster recovery testing
- Plan updates after incidents
- Testing results documented

#### Evaluation (§164.308(a)(8))

**Periodic Evaluation**
- Annual HIPAA compliance assessment
- Internal audits quarterly
- External penetration testing annually
- Evaluation findings documented

#### Business Associate Contracts (§164.308(b)(1))

**Written Contract Required**
- BAA signed with all customers handling PHI
- BAA includes all required provisions
- Sub-contractors also sign BAAs
- Contract retention: 6 years after termination

**BAA Requirements**:
- Description of permitted uses of PHI
- Prohibition of unauthorized use/disclosure
- Safeguards to protect PHI
- Breach reporting obligations
- Return or destruction of PHI upon termination

### Physical Safeguards (§164.310)

#### Facility Access Controls (§164.310(a))

**Contingency Operations**
- Alternate processing site available
- Multi-region deployment capability
- Failover procedures documented

**Facility Security Plan**
- Cloud provider physical security (AWS/GCP/Azure)
- Data center access controls
- Video surveillance
- Visitor logs maintained

**Access Control and Validation**
- Badge access to facilities
- Biometric authentication for critical areas
- Access logs reviewed monthly

**Maintenance Records**
- Hardware maintenance logged
- Repairs documented
- Disposal procedures for equipment

#### Workstation Use (§164.310(b))

**Proper Use**
- Workstation use policy defined
- Screen privacy filters
- Clean desk policy
- Auto-lock after 15 minutes idle

#### Workstation Security (§164.310(c))

**Physical Safeguards**
- Laptop encryption (full disk)
- Cable locks for workstations
- Secure disposal procedures
- No PHI on mobile devices without encryption

#### Device and Media Controls (§164.310(d))

**Disposal**
- Secure erasure before disposal
- Physical destruction if necessary
- Disposal documented
- Certificate of destruction retained

**Media Re-use**
- Sanitization before re-use
- Verification of data removal
- Re-use testing procedures

**Accountability**
- Asset inventory maintained
- Serial numbers tracked
- Check-in/check-out procedures
- Loss reporting process

**Data Backup and Storage**
- Encrypted backup storage
- Off-site backup location
- Access controls on backups
- Backup encryption keys secured

### Technical Safeguards (§164.312)

#### Access Control (§164.312(a))

**Unique User Identification (Required)**
- Unique username for each user
- No shared accounts
- Service accounts closely monitored
- User IDs never reused

**Emergency Access Procedure (Required)**
- Break-glass access procedures
- Emergency access logged and reviewed
- Justification required
- Time-limited emergency access

**Automatic Logoff (Addressable)**
- Web session timeout: 15 minutes
- API token expiration: 24 hours
- Idle session termination
- User notified before logoff

**Encryption and Decryption (Addressable)**
- AES-256-GCM encryption at rest
- TLS 1.3 encryption in transit
- Encryption keys in HSM or cloud KMS
- Key rotation every 90 days

#### Audit Controls (§164.312(b))

**Audit Logging**
- All PHI access logged (who, what, when, where)
- Immutable audit trail
- Tamper detection (HMAC signatures)
- Log retention: 6 years minimum

**Logged Events**:
- User authentication (success/failure)
- PHI access (read/write/delete)
- Permission changes
- Security incidents
- Configuration changes
- Backup/restore operations

#### Integrity (§164.312(c))

**Mechanism to Authenticate ePHI**
- Digital signatures for critical data
- Checksums for file integrity
- Database constraints
- Referential integrity enforced

**Protection from Tampering**
- Audit log signatures (HMAC)
- Immutable tables (no UPDATE/DELETE)
- Version control for documents
- Tamper detection alerts

#### Person or Entity Authentication (§164.312(d))

**Authentication**
- Strong password requirements
- Multi-factor authentication (TOTP, WebAuthn)
- Biometric options (WebAuthn)
- Certificate-based authentication for services

#### Transmission Security (§164.312(e))

**Integrity Controls**
- TLS 1.3 for all transmissions
- Message authentication codes
- Checksums for data integrity
- Replay attack prevention

**Encryption**
- End-to-end encryption for PHI
- TLS 1.3 minimum version
- Strong cipher suites only
- Certificate validation enforced

## HIPAA Privacy Rule

### Notice of Privacy Practices

**Required Elements**:
- How PHI may be used and disclosed
- Individual rights regarding PHI
- Covered entity obligations
- Contact for privacy questions
- Effective date and signature

### Individual Rights

**Right to Access (§164.524)**
- Provide PHI within 30 days
- Electronic format if requested
- Reasonable fees allowed
- Access denial process defined

**Right to Amend (§164.526)**
- Allow corrections within 60 days
- Denial reasons documented
- Amendment appended to record

**Right to Accounting of Disclosures (§164.528)**
- Track PHI disclosures
- Provide accounting within 60 days
- Retention: 6 years

**Right to Request Restrictions (§164.522)**
- Allow requests for restrictions
- Not required to agree (except out-of-pocket full payment)
- Restrictions documented if agreed

### Minimum Necessary

**Minimum Necessary Standard**
- Limit PHI use/disclosure to minimum needed
- Role-based access implements minimum necessary
- Queries limited to necessary fields
- Disclosure review process

### De-Identification

**Safe Harbor Method**
- Remove all 18 identifiers
- No actual knowledge of re-identification
- De-identified data not subject to HIPAA

**Expert Determination Method**
- Statistical analysis by expert
- Very small risk of re-identification
- Expert opinion documented

## HIPAA Breach Notification Rule

### Breach Assessment

**4-Factor Risk Assessment**:
1. Nature and extent of PHI involved
2. Who unauthorized person is
3. Whether PHI was actually acquired/viewed
4. Extent to which risk mitigated

### Notification Requirements

**Individual Notification (§164.404)**
- Within 60 days of discovery
- Written notification by first-class mail
- Electronic notification if preferred
- Substitute notice if contact information insufficient

**Media Notification (§164.406)**
- If breach affects 500+ individuals in state/jurisdiction
- Prominent media outlet notification
- Within 60 days of discovery

**HHS Notification (§164.408)**
- 500+ individuals: Immediately (concurrent with individual notification)
- <500 individuals: Annual log submitted
- Breach log maintained

### Breach Documentation

**Required Documentation**:
- Discovery date
- Affected individuals count
- Nature of breach
- PHI involved
- Cause of breach
- Risk assessment
- Mitigation actions
- Notifications sent
- Retention: 6 years minimum

## HIPAA Compliance Checklist

### Administrative
- [x] Business Associate Agreements signed
- [x] HIPAA Security Officer designated
- [x] Privacy Officer designated
- [x] Risk assessment conducted
- [x] Security policies documented
- [x] Workforce training program
- [x] Sanction policy defined
- [x] Incident response procedures

### Physical
- [x] Cloud provider physical security verified
- [x] Workstation security policy
- [x] Device disposal procedures
- [x] Backup storage security
- [x] Facility access controls

### Technical
- [x] Unique user IDs
- [x] Emergency access procedures
- [x] Automatic logoff (15 minutes)
- [x] Encryption at rest (AES-256)
- [x] Encryption in transit (TLS 1.3)
- [x] Audit logging (immutable, 6-year retention)
- [x] Multi-factor authentication
- [x] Transmission security (TLS 1.3)

### Privacy
- [x] Notice of Privacy Practices
- [x] Individual rights procedures
- [x] Minimum necessary access
- [x] Breach notification procedures
- [x] De-identification support

## HIPAA Training Requirements

### Initial Training
- All workforce members handling PHI
- Within 30 days of hire
- Covers HIPAA basics, privacy, security
- Training documented

### Ongoing Training
- Annual refresher training
- Training on policy changes
- Incident-based training
- Training records retained 6 years

### Training Topics
- PHI definition and examples
- Permitted uses and disclosures
- Minimum necessary standard
- Patient rights
- Breach notification
- Security awareness
- Incident reporting

## Contact Information

**HIPAA Security Officer**
Email: security@kapok.io

**HIPAA Privacy Officer**
Email: privacy@kapok.io

**Breach Reporting (24/7)**
Email: security@kapok.io
Phone: [To be configured]

## Document Control

- **Version**: 1.0
- **Last Updated**: 2026-01-29
- **Next Review**: 2026-04-29 (Quarterly)
- **Owner**: HIPAA Security Officer
- **Classification**: Internal
