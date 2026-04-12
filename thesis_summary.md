
# A Mobile-First Smart Document Management Platform for Personal and Small-Business Records

**Author:** Mohammed Bin Saif Bin Khalfan Al Jahwari
**Student ID:** 20-0201
**Degree:** Bachelor of Science in Computer Science
**Institution:** German University of Technology in Oman
**Date:** Spring 2026

## Abstract

Important documents are often stored in places that were never meant to function as a reliable records system. Personal users keep warranties, certificates, and identity papers in phone galleries, messaging applications, or email threads, while small businesses scatter invoices, contracts, and supplier records across shared drives and ad hoc folders. The result is a familiar problem: documents are hard to retrieve, key dates are missed, and there is little control over access, history, or accountability.

This thesis presents the design of a mobile-first smart document management platform for personal and small-business records. The proposed system is conceived as a SaaS product with two primary interfaces: a web application and a mobile application. Its main functions are secure document storage, OCR-first text extraction using Mistral OCR, metadata enrichment, pgvector-based semantic retrieval for a RAG workflow, and reminder workflows for time-sensitive records. Web upload and phone-camera capture are treated as first-class use cases, while channels such as email forwarding and WhatsApp sharing are considered future extensions rather than core MVP requirements.

The design adopts a selective microservice architecture composed of a Go-based core platform service, a Python-based document-processing service, and an asynchronous worker layer. All uploaded files — including photographs, scanned pages, and born-digital PDFs — pass through a common OCR-first pipeline centered on Mistral OCR so that the system maintains one searchable text path instead of multiple inconsistent extraction branches. Search applies metadata filters and semantic retrieval through pgvector, where each stored chunk carries grounding metadata such as document identity and page location for the later RAG step. Security is addressed through tenant-aware access control, role-based permissions, audit logging, encrypted storage, and short-lived download URLs.

The contribution of the thesis is not a production implementation but a defensible system design. It brings together product scope definition, literature-informed architectural choices, a database and API design, and a practical strategy for multilingual document processing in a form suitable for staged implementation.

## Table of Contents

1. [Introduction](#introduction)
   - 1.1 Background
   - 1.2 Problem Statement
   - 1.3 Project Motivation
   - 1.4 Objectives
   - 1.5 Research Questions
   - 1.6 Scope and Limitations
   - 1.7 Significance of the Study
   - 1.8 Organization of the Thesis
   - 1.9 Summary of the Chapter

2. [Literature Review](#literature-review)
   - 2.1 Document Management Systems
   - 2.2 Security and Governance in Document Repositories
   - 2.3 OCR and Intelligent Document Processing
   - 2.4 Arabic OCR and Multilingual Challenges
   - 2.5 Search and Retrieval in Document Platforms
   - 2.6 Embedding Choice and Provider Flexibility
   - 2.7 Architecture Patterns for a SaaS Document Platform
   - 2.8 Research Gap and Design Implications
   - 2.9 Summary of the Chapter

3. [Research Methodology](#research-methodology)
   - 3.1 Research Design
   - 3.2 Sources of Evidence
   - 3.3 Decision Criteria
   - 3.4 Design Procedure
   - 3.5 Validation Strategy
   - 3.6 Product Vision
   - 3.7 Primary Users and Usage Scenarios
   - 3.8 Functional Requirements
   - 3.9 Non-functional Requirements
   - 3.10 MVP Scope and Deferred Features
   - 3.11 Evaluation Targets
   - 3.12 Methodological Limitations
   - 3.13 Summary of the Chapter

4. [Implementation](#implementation)
   - 4.1 Architectural Overview
   - 4.2 Architectural Rationale
   - 4.3 Major Runtime Components
   - 4.4 Core Workflows
   - 4.5 Data Model and Storage Design
   - 4.6 API Surface and Client Boundary
   - 4.7 Document Processing and Retrieval Strategy
   - 4.8 Technology Stack
   - 4.9 Summary of the Chapter

5. [Conclusion and Future Work](#conclusion-and-future-work)
   - 5.1 Summary of the Study
   - 5.2 Key Findings
   - 5.3 Research Questions Revisited
   - 5.4 Limitations and Risks
   - 5.5 Future Work
   - 5.6 Contributions
   - 5.7 Concluding Remarks

## List of Tables

- Table 4.1: Recommended technology stack for the proposed platform
- Table 5.1: Primary project risks and mitigation strategies
- Table A.1: Supplementary comparison of OCR and document-processing options
- Table A.2: Supplementary comparison of embedding options

## List of Figures

- Figure 4.1: System architecture showing the core platform boundary, asynchronous OCR-first processing path, and shared storage services.
- Figure 4.2: OCR-first ingestion workflow with parallel document-processing and reminder workers.
- Figure 4.3: pgvector-based retrieval workflow for the RAG application.
- Figure 4.4: Entity-relationship diagram for the proposed platform schema.

## List of Symbols and Abbreviations

| Term | Meaning |
| --- | --- |
| API | Application Programming Interface |
| EDMS | Electronic Document Management System |
| LLM | Large Language Model |
| MVP | Minimum Viable Product |
| OIDC | OpenID Connect |
| OCR | Optical Character Recognition |
| PDF | Portable Document Format |
| RAG | Retrieval-Augmented Generation |
| RBAC | Role-Based Access Control |
| SaaS | Software as a Service |
| SSO | Single Sign-On |
| VLM | Vision-Language Model |

## Chapter 1: Introduction

### 1.1 Background

Digital documents have become everyday records. A warranty card, an insurance form, a passport copy, an invoice, or a signed contract may start life as a paper sheet, a PDF received by email, or a photograph taken on a phone. In both personal and business settings, these materials carry operational value: they prove ownership, support compliance, document payment, and record obligations. Yet the tools used to keep them are usually fragmented. A user may save one file in a cloud drive, another in WhatsApp, another in a phone gallery, and several more in email attachments.

### 1.2 Problem Statement

Important documents are stored in tools that were not designed for long-term records management. Invoices sit in email, contracts in shared folders, warranties as phone images, and identity documents in scanning applications. This fragmentation produces four concrete problems: documents become hard to retrieve, critical dates are missed, Arabic and mixed-language records degrade search quality, and lightweight storage tools provide no meaningful access control or audit trail.

### 1.3 Project Motivation

The motivation for the project comes from a practical observation: people often need a document at the exact moment they cannot find it. A student may need a certificate, a family may need an insurance paper, and a small business owner may need an invoice, a supplier agreement, or a renewal notice. In many cases, the document exists, but it is buried in the wrong application or saved under an unhelpful filename.

### 1.4 Objectives

The primary objective of this thesis is to produce a coherent system design for a mobile-first smart document management platform. This objective is divided into the following specific goals:

1. Define a suitable product scope for a thesis-scale SaaS document platform.
2. Identify the main user groups, use cases, and MVP boundaries.
3. Justify an architecture that separates transactional platform logic from document-processing workloads.
4. Design a document pipeline that combines OCR, metadata extraction, and pgvector-based bilingual retrieval.
5. Specify a database and API structure that supports secure storage, search, and reminder workflows.
6. Integrate security, observability, and deployment concerns into the design from the start.

### 1.5 Research Questions

The thesis is guided by the following research questions:

1. What product scope is appropriate for a mobile-first document management platform serving personal and small-business users?
2. Which architecture best balances simplicity, maintainability, and future scalability for this problem?
3. How should OCR, metadata extraction, and Arabic-English retrieval be combined in one practical processing strategy?
4. Which technology choices are suitable for an MVP without creating avoidable vendor or operational lock-in?
5. How should security, access control, and operational visibility be incorporated into the design?

### 1.6 Scope and Limitations

This thesis is design-oriented rather than implementation-oriented. It proposes and justifies a system architecture, data model, API boundary, and processing strategy, but it does not deliver a production system. The thesis assumes a SaaS deployment model and treats the web application and mobile application as the primary user interfaces.

### 1.7 Significance of the Study

The significance of the study lies in its attempt to connect several concerns that are often discussed separately. Document-management literature emphasizes governance and records control. OCR and document-AI literature emphasizes extraction quality. Retrieval literature emphasizes semantic search, grounding, and provenance. Software architecture literature emphasizes service boundaries and operational trade-offs. This thesis brings these strands together around a concrete product idea: a platform that begins with the simple act of capturing a document on a phone or uploading it from the web, then turns that document into a manageable and searchable record.

### 1.8 Organization of the Thesis

The thesis is organized into five chapters. Chapter 2 reviews the literature on document management systems, OCR, multilingual retrieval, and architecture patterns. Chapter 3 explains the research methodology and defines the product vision, requirements, and MVP scope. Chapter 4 presents the system design and architecture, including the data model, document processing strategy, and technology stack. Chapter 5 concludes the thesis, discusses limitations, and outlines future work.

### 1.9 Summary of the Chapter

This chapter introduced the problem addressed by the thesis, explained the motivation behind the proposed platform, stated the objectives and research questions, and clarified the study scope.

## Chapter 2: Literature Review

### 2.1 Document Management Systems

Document management is not a new problem, but its practical context has changed. Earlier electronic document and records management studies focused mainly on institutional archives, administrative offices, and formal records-control environments. In that literature, the key concerns are consistent classification, access control, retrieval, retention, and system integration.

### 2.2 Security and Governance in Document Repositories

Security is consistently treated in the literature as a central design requirement rather than an optional enhancement. Secure document systems must address who can view a file, who can change metadata, how actions are recorded, and how sensitive data is protected in storage and transit.

### 2.3 OCR and Intelligent Document Processing

OCR remains the foundational technology behind most practical document-intelligence systems. Managed platforms from Google Document AI, AWS Textract, Azure OCR, and Azure Document Intelligence present OCR as one part of a broader processing pipeline that may include layout detection, table parsing, key-value extraction, and downstream automation.

### 2.4 Arabic OCR and Multilingual Challenges

Arabic document processing remains one of the most difficult parts of the problem domain. Arabic script introduces challenges related to cursive writing, right-to-left flow, font variation, diacritics, and numeral handling.

### 2.5 Search and Retrieval in Document Platforms

Retrieval in document systems usually begins with structure. Metadata fields such as document type, issuer, issue date, expiry date, and tags allow transparent filtering, but they are only as good as the extraction process that produced them.

### 2.6 Embedding Choice and Provider Flexibility

Embedding models matter because they affect retrieval quality, cost, language support, latency, and long-term portability. Managed APIs such as Google’s gemini-embedding-001 offer documented multilingual retrieval support and convenient deployment.

### 2.7 Architecture Patterns for a SaaS Document Platform

The architecture literature does not offer a simple rule that microservices are always better. Studies on migration and refactoring show that service decomposition can improve isolation and scaling when it reflects genuine business or runtime boundaries, but it can also introduce coordination overhead, deployment burden, and anti-patterns when applied too early or too finely.

### 2.8 Research Gap and Design Implications

The literature reviewed in this chapter reveals a useful gap. Document-management studies explain governance, integration, and secure handling. OCR and document-understanding research explain extraction choices. Retrieval research explains dense multilingual search and grounded evidence use. Architecture research explains service-boundary trade-offs.

### 2.9 Summary of the Chapter

This chapter reviewed the literature relevant to the thesis from five perspectives: document-management systems, security and governance, OCR and document understanding, multilingual retrieval, and software architecture.

## Chapter 3: Research Methodology

### 3.1 Research Design

This thesis follows a design-oriented comparative methodology. The goal is not to prove the performance of a finished implementation, but to construct and justify a system design that is academically defensible and practically implementable.

### 3.2 Sources of Evidence

Three sources of evidence are used throughout the thesis. First, academic literature is used to understand document-management systems, OCR, multilingual retrieval, and software architecture. Second, official vendor documentation is used where current implementation details matter, especially for OCR services, embedding APIs, and database capabilities. Third, model cards and project documentation are used to assess open alternatives.

### 3.3 Decision Criteria

The main design decisions in the thesis are evaluated against a common set of criteria:

1. Problem fit
2. Maintainability
3. Scalability relevance
4. Multilingual readiness
5. Security and governance
6. Incremental delivery

### 3.4 Design Procedure

The design process used in the thesis has five steps:

1. Define the product problem in practical terms.
2. Review the literature to identify the main technical and organizational issues.
3. Assemble realistic alternatives for each major design area.
4. Compare these alternatives using the criteria stated above.
5. Select an MVP-oriented design, while features that are attractive but not yet justified are explicitly deferred.

### 3.5 Validation Strategy

Because the thesis is not an implementation study, validation is carried out through internal coherence and future measurability rather than experimental deployment.

### 3.6 Product Vision

The proposed product is a SaaS document-management platform designed primarily for a web application and a mobile application. Its purpose is to help users capture, store, organize, search, and monitor important records without depending on scattered folders, email threads, or messaging history.

### 3.7 Primary Users and Usage Scenarios

The requirements are shaped by four main user groups:

- Personal users who store warranties, receipts, identity records, insurance papers, and certificates.
- Small-business users who need searchable invoices, quotations, supplier documents, and contracts.
- Operational staff who upload, classify, or correct documents on behalf of an organization.
- Administrative reviewers who need access control, audit visibility, and document history.

### 3.8 Functional Requirements

#### 3.8.1 Document Capture and Ingestion

The platform must accept uploads in PDF and common image formats. Users should be able to submit files from the web interface or directly from the mobile application after taking a photo with a phone camera.

#### 3.8.2 Organization and Metadata Management

Users must be able to organize documents through folders, document types, tags, and editable metadata.

#### 3.8.3 Search and Retrieval

The platform must support metadata-guided semantic retrieval. Users should be able to apply filters such as document type or date and then search over chunk embeddings stored in pgvector.

#### 3.8.4 Reminders and Notifications

The system must support reminder workflows for time-sensitive records such as expiry dates, renewal dates, and due dates.

#### 3.8.5 Security and Administrative Control

Because the platform manages sensitive material, it must support tenant-aware access control, organization membership, role-based permissions, and audit logging.

### 3.9 Non-functional Requirements

The system design is also shaped by several non-functional requirements:

1. Maintainability by a small project team
2. Reliable and retry-safe document processing
3. Explainable search layer
4. Privacy and secure storage
5. Staged growth

### 3.10 MVP Scope and Deferred Features

The MVP includes secure upload, OCR-first processing, editable metadata, bilingual pgvector-based retrieval, reminders, role-based access control, and basic operational monitoring.

### 3.11 Evaluation Targets

Although the thesis does not present a live implementation, it defines concrete targets for later evaluation:

- OCR should be assessed on representative Arabic and English documents.
- Retrieval should be evaluated through precision and recall on a bilingual document set.
- Metadata extraction should be checked for the main document categories.
- Reminder performance should be evaluated by how accurately the system identifies and schedules date-based events.

### 3.12 Methodological Limitations

The main limitation of the methodology is that some technology choices are assessed through published results and documentation rather than direct experimentation on a project-specific dataset.

### 3.13 Summary of the Chapter

This chapter described the methodology used to build the thesis: a comparative, design-oriented approach supported by literature, documentation, and explicit decision criteria.

## Chapter 4: Implementation

### 4.1 Architectural Overview

The proposed platform follows a selective microservice architecture centered on a core platform service and an asynchronous processing layer built around an OCR-first workflow.

### 4.2 Architectural Rationale

The architectural decision is driven by workload differences. API requests for login, listing documents, loading metadata, or searching the repository should respond quickly and predictably. OCR, document enrichment, retrieval preparation, and reminder extraction behave differently: they are slower, more compute-heavy, and better suited to asynchronous execution.

### 4.3 Major Runtime Components

#### 4.3.1 Core Platform Service

The core service is the system’s main application boundary. It owns users, organizations, permissions, document records, metadata edits, reminder configuration, and the public API.

#### 4.3.2 Document-Processing Service

The document-processing service consumes OCR output and handles layout-aware parsing, document classification, metadata extraction, retrieval-chunk preparation, and persistence of processed outputs.

#### 4.3.3 Worker Layer

The worker layer handles asynchronous operational tasks such as reminder-date extraction, reminder scheduling, notification delivery, retry logic, and lower-priority background jobs.

#### 4.3.4 Shared Data Layer

The platform stores document binaries in object storage and stores transactional and retrieval data in PostgreSQL.

### 4.4 Core Workflows

#### 4.4.1 Document Ingestion

A user uploads a file through the web application or mobile application. The core platform stores the original binary, creates the initial document and version records, and publishes a processing job.

#### 4.4.2 Search and Retrieval

Every query is first constrained by tenant and permission checks. The system then applies metadata filters and runs semantic search over chunk embeddings stored in PostgreSQL with pgvector.

### 4.5 Data Model and Storage Design

The proposed schema is centered on tenant-aware records. At the root of the schema are tenants, organizations, users, and memberships. These entities define isolation and access boundaries.

### 4.6 API Surface and Client Boundary

The platform exposes a versioned REST API that acts as the shared boundary for the web and mobile clients.

### 4.7 Document Processing and Retrieval Strategy

#### 4.7.1 OCR-First Ingestion

The platform adopts an OCR-first strategy for all uploaded documents. This includes photographed pages, scanned PDFs, and born-digital PDFs.

#### 4.7.2 Processing Stages

The proposed document-processing pipeline has five stages:

1. OCR-based text extraction
2. Metadata extraction and normalization
3. Page-aware chunk preparation for retrieval
4. Embedding generation for each chunk
5. Persistence of chunk text, embeddings, and metadata for retrieval

#### 4.7.3 OCR Technology Choice

For the MVP, Mistral OCR 2503 is selected as the primary OCR service.

#### 4.7.4 Metadata Extraction Strategy

Metadata extraction is designed as a layered process rather than a single model call.

#### 4.7.5 Chunking and Retrieval Metadata

The retrieval unit is a chunk rather than a whole document. Each chunk stores the OCR text segment together with the metadata needed later by the application.

#### 4.7.6 pgvector Retrieval and RAG

The retrieval layer is intentionally narrow. After tenant and permission checks, the system applies any metadata filters and then runs semantic search over chunk embeddings stored in PostgreSQL with pgvector.

#### 4.7.7 Multilingual Retrieval for Arabic and English

Arabic-English support is treated as a first-class requirement, not as a cosmetic extension.

#### 4.7.8 Role of the LLM Layer

Large language models are useful in this design, but they are deliberately constrained.

### 4.8 Technology Stack

| Layer | Recommendation | Rationale |
| --- | --- | --- |
| Frontend | Next.js with TypeScript | Mature web framework with a strong ecosystem for application development |
| Core backend | Go service | Suitable for a maintainable API layer and operationally simple deployment |
| OCR service | Mistral OCR 2503 | OCR-focused managed service that fits the MVP ingestion boundary and API-first integration model |
| Document processing and RAG orchestration | Python | Appropriate for OCR integration, chunk preparation, embeddings, and retrieval orchestration |
| Relational and vector storage | PostgreSQL with pgvector | Supports transactional data, chunk metadata, and vector retrieval in one system |
| Object storage | S3-compatible object store | Separates binaries from metadata and search state |
| Queue | RabbitMQ or equivalent managed broker | Supports asynchronous processing and retries |
| Observability | Sentry + Grafana + OpenTelemetry | Covers errors, metrics, logs, and traces |
| Authentication | Custom authentication service with SSO support | Preserves control over tenant and organization logic while supporting enterprise identity requirements |

### 4.9 Summary of the Chapter

This chapter presented the implementation design of the proposed platform. It justified the selective microservice structure, described the main runtime components, explained the ingestion and retrieval workflows, outlined the core data model and API boundary, discussed the document-processing and retrieval strategy, and summarized the technology stack.

## Chapter 5: Conclusion and Future Work

### 5.1 Summary of the Study

This thesis presented the design of a mobile-first smart document management platform for personal and small-business records.

### 5.2 Key Findings

Five outcomes stand out from the thesis:

1. The product scope was narrowed to what is genuinely important for an MVP.
2. The literature review showed that OCR remains a practical canonical text layer.
3. The study justified a simplified pgvector-based retrieval strategy.
4. It argued for a restrained service split.
5. It defined web and mobile applications as the primary client channels.

### 5.3 Research Questions Revisited

The research questions posed in Chapter 1 are answered as follows:

- RQ1: The appropriate product scope is mobile-first document capture with OCR-backed ingestion, metadata-guided semantic search, reminders, and role-based access control.
- RQ2: A selective microservice architecture with a core platform service, document-processing service, and worker layer best balances simplicity, maintainability, and future scalability.
- RQ3: OCR, metadata extraction, and Arabic-English retrieval should be combined through an OCR-first pipeline centered on Mistral OCR, layered metadata extraction, and pgvector-backed bilingual semantic search.
- RQ4: An MVP stack centered on Go, Python, PostgreSQL with pgvector, Mistral OCR, and S3-compatible object storage provides suitable choices without creating avoidable vendor or operational lock-in.
- RQ5: Security, access control, and operational visibility should be incorporated through tenant-scoped data models, RBAC, audit logging, encrypted storage, and disciplined observability.

### 5.4 Limitations and Risks

This thesis is a design study, not an implementation study. For that reason, several important questions remain open.

### 5.5 Future Work

The first priority for future work is implementation and evaluation. A representative bilingual corpus should be assembled from the main document categories discussed in the thesis.

### 5.6 Contributions

The main contribution of this thesis is a coherent, defensible system design for a mobile-first SaaS document management platform.

### 5.7 Concluding Remarks

The main argument of the thesis is that smart document management should begin with disciplined document handling, not with unnecessary complexity.

## Appendix A: Supplementary Vendor and Model Comparison Tables

### Table A.1: Supplementary comparison of OCR and document-processing options

| Option | Core capability | Layout / forms | Arabic posture |
| --- | --- | --- | --- |
| Google Document AI | Managed document-intelligence platform with OCR, parsers, and workflow-oriented processors | Broad document-processing scope with specialized processors | Must still be benchmarked on Arabic corpus, but the platform is explicitly multilingual |
| Google Document AI Enterprise OCR | Managed OCR with page quality, language hints, rotation correction, OCR add-ons, native text options | Strong layout and OCR add-ons | Must still be benchmarked on Arabic corpus, but multilingual OCR controls are explicit |
| AWS Textract | OCR, forms, tables, queries, expense, ID, lending | Strong form and table capabilities | Requires corpus validation for Arabic-specific quality |
| Azure OCR | Managed OCR for images with lightweight synchronous processing | Better suited to simpler OCR scenarios than full document pipelines | Arabic suitability should be validated against document-heavy use cases |
| Azure Document Intelligence | Layout, text, tables, roles, custom extraction, custom classification | Strong structural parsing | Arabic support should be validated on thesis corpus |
| Mistral OCR 2503 (Selected) | OCR-focused multimodal model for text and structured content extraction from documents | Strong API-native document extraction posture | Chosen for the thesis design, but Arabic performance still requires independent benchmarking |
| Tesseract | Open-source OCR engine with LSTM models | Limited by surrounding pipeline tooling | Broad language support through trained data |
| PaddleOCR 3.0 | Open-source OCR and document parsing toolkit | Stronger document parsing support than classical OCR engines | Explicit multilingual positioning |
| DeepSeek-OCR | OCR-focused research model line emphasizing context-aware document compression and recognition | Research-oriented rather than production-ready service evidence | Arabic posture is unclear and should not be assumed without direct evaluation |
| GLM-OCR | OCR-focused technical-report model line positioned for document understanding workflows | Research signal for fast-moving OCR model development | Arabic posture is unclear and should not be assumed without direct evaluation |

### Table A.2: Supplementary comparison of embedding options

| Model | Claimed strengths | Context |
| --- | --- | --- |
| Gemini Embedding 2 / gemini-embedding-001 | Google documents multilingual and retrieval-oriented usage with task-specific embedding modes | 2048-token sequence length |
| OpenAI / text-embedding-3-large | Strong managed API option with multilingual performance claims and broad ecosystem adoption for retrieval workflows | 8192-token max input |
| multilingual-e5-large | Strong multilingual benchmark evidence and broad language support | 512-token truncation in model card usage notes |
| BGE-M3 | Multilingual, multi-function, and multi-granularity design suitable for short and long retrieval units | Attractive for document corpora mixing metadata snippets and OCR chunks |
| Qwen3 Embedding | Foundation-model-based embedding family aimed at stronger retrieval pipelines | Relevant as a candidate multilingual embedding model for chunk retrieval |
| Arctic-Embed 2.0 | Strong multilingual retrieval claims with compression-friendly design | Long-context support in model card |
