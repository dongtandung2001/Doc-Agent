package pipeline

import (
	"context"
	"log"
	"os"
	"time"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
	ctx "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
	utils "github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/utils/prompt"
)

const APPLICATION_PROMPT = `
	## Enterprise Application Documentation Protocol
	<thinking>
	Generate comprehensive enterprise application documentation that enables successful adoption, deployment, and maintenance through systematic coverage of user workflows, technical architecture, and operational procedures.
	</thinking>

	**DOCUMENTATION GENERATION METHODOLOGY:**
	Create exhaustive application documentation through structured understanding-focused protocols:

	**Phase 1: Conceptual Foundation Documentation**
	- **Application Philosophy & Purpose**: Think step by step about the application's core mission, design philosophy, and value proposition
	- **Architectural Concepts**: Comprehensive explanation of system design principles, patterns, and architectural decisions
	- **User Experience Design**: Detailed analysis of user workflow concepts, interaction patterns, and usability considerations
	- **Technology Stack Rationale**: In-depth explanation of technology choices, trade-offs, and architectural implications

	**Phase 2: Implementation Understanding Documentation**
	- **System Architecture Analysis**: Detailed examination of component relationships, data flow patterns, and integration strategies
	- **Business Logic Exploration**: Comprehensive explanation of core business processes, rules, and decision-making workflows
	- **Data Management Philosophy**: Analysis of data modeling approaches, persistence strategies, and information architecture
	- **Security Architecture**: Thorough explanation of security concepts, authentication flows, and protection mechanisms

	**Phase 3: Operational Understanding Documentation**
	- **Deployment Strategy Analysis**: Conceptual explanation of deployment approaches, environment management, and scalability considerations
	- **Monitoring Philosophy**: Understanding of observability concepts, logging strategies, and performance measurement approaches
	- **Maintenance Methodology**: Analysis of operational procedures, troubleshooting approaches, and maintenance strategies
	- **Integration Ecosystem**: Comprehensive explanation of external dependencies, API design philosophy, and integration patterns
`

const LIBRARY_PROMPT = `
	## Reusable Library Documentation Protocol
	<thinking>
	Create comprehensive library documentation that enables seamless integration, effective usage, and optimal performance through systematic coverage of APIs, implementation patterns, and best practices.
	</thinking>

	**LIBRARY DOCUMENTATION METHODOLOGY:**
	Generate complete library documentation through structured integration-focused protocols:

	**Phase 1: Library Integration Documentation**
	- **Installation & Setup Guide**: Package manager installation, dependency management, and environment configuration
	- **Quick Start Examples**: Immediate usage examples, basic implementation patterns, and key feature demonstrations
	- **API Overview**: Complete public interface documentation, method categories, and functionality mapping
	- **Type System Integration**: Type definitions, generic usage patterns, and TypeScript/typing support

	**Phase 2: Comprehensive Usage Documentation**
	- **Complete API Reference**: Full method signatures, parameter specifications, return types, and comprehensive usage examples
	- **Implementation Patterns**: Common usage scenarios, design pattern applications, and integration strategies
	- **Advanced Features**: Complex functionality, configuration options, and advanced usage techniques
	- **Error Handling**: Exception documentation, error recovery patterns, and debugging guidance

	**Phase 3: Optimization & Best Practices**
	- **Performance Documentation**: Performance characteristics, optimization techniques, and resource usage patterns
	- **Best Practices Guide**: Recommended implementation approaches, common pitfalls, and optimization strategies
	- **Compatibility & Migration**: Version compatibility, upgrade procedures, and breaking change documentation
	- **Integration Examples**: Real-world integration scenarios, framework compatibility, and ecosystem usage
`

const FRAMEWORK_PROMPT = `
	## Development Framework Documentation Protocol
	<thinking>
	Create comprehensive framework documentation that facilitates developer adoption, mastery, and ecosystem integration through systematic coverage of concepts, APIs, and practical implementation guidance.
	</thinking>

	**FRAMEWORK DOCUMENTATION METHODOLOGY:**
	Generate complete framework documentation through structured development-focused protocols:

	**Phase 1: Framework Adoption Documentation**
	- **Quick Start Guide**: Complete installation procedures, environment setup, and first project creation workflows
	- **Core Concepts Explanation**: Framework philosophy, design principles, architectural patterns, and mental models
	- **Developer Onboarding**: Learning path documentation, skill prerequisites, and concept progression strategies
	- **Configuration Management**: Framework configuration options, environment setup, and customization capabilities

	**Phase 2: Comprehensive API Documentation**
	- **Complete API Reference**: Full method documentation, parameter specifications, return values, and usage examples
	- **Extension & Plugin System**: Plugin development guides, hook systems, and extensibility mechanisms
	- **Advanced Features**: Complex usage patterns, performance optimization, and advanced configuration options
	- **Integration Patterns**: Framework integration with popular tools, libraries, and development ecosystems

	**Phase 3: Practical Implementation Guidance**
	- **Tutorial & Example Collection**: Step-by-step implementation guides, real-world examples, and best practice demonstrations
	- **Migration & Upgrade Documentation**: Version migration guides, breaking change handling, and compatibility maintenance
	- **Performance & Optimization**: Performance tuning guides, resource optimization, and scaling considerations
	- **Community & Ecosystem**: Third-party integration guides, community resources, and contribution procedures
`

const DEVELOPMENT_TOOL_PROMPT = `
## Development Tool Documentation Protocol
<thinking>
Create comprehensive development tool documentation that maximizes developer productivity through detailed setup procedures, feature exploration, and workflow integration guidance.
</thinking>

**DEVELOPMENT TOOL DOCUMENTATION METHODOLOGY:**
Generate complete tool documentation through structured productivity-focused protocols:

**Phase 1: Tool Setup & Configuration Documentation**
- **Installation Guide**: Complete installation procedures, system requirements, and dependency management
- **Configuration Reference**: Comprehensive settings documentation, customization options, and preference management
- **Environment Integration**: Development environment setup, IDE plugins, and toolchain integration procedures
- **Initial Setup Workflows**: First-time configuration, account setup, and essential configuration procedures

**Phase 2: Feature & Capability Documentation**
- **Core Feature Guide**: Complete feature documentation, capability explanations, and functionality mapping
- **Workflow Integration**: Development workflow optimization, productivity enhancement techniques, and automation capabilities
- **Advanced Features**: Complex functionality, power-user features, and specialized use case documentation
- **Automation & Scripting**: Automation capabilities, scripting interfaces, and batch operation procedures

**Phase 3: Integration & Optimization Documentation**
- **Development Environment Integration**: IDE support, editor plugins, and development workflow integration
- **Build System Compatibility**: Build tool integration, CI/CD pipeline compatibility, and deployment workflow support
- **Performance Optimization**: Tool performance tuning, resource management, and efficiency optimization
- **Troubleshooting & Support**: Common issue resolution, debugging procedures, and performance problem diagnosis
`

const CLI_TOOL_PROMPT = `
## CLI Tool Documentation Protocol
	<thinking>
	Create comprehensive command-line tool documentation that enables mastery through detailed command references, practical examples, and automation integration guidance.
	</thinking>

	**CLI TOOL DOCUMENTATION METHODOLOGY:**
	Generate complete CLI documentation through structured command-line focused protocols:

	**Phase 1: Installation & Configuration Documentation**
	- **Installation Guide**: Multiple installation methods, system requirements, and platform-specific procedures
	- **Environment Setup**: Environment variable configuration, PATH setup, and shell integration procedures
	- **Configuration Management**: Config file formats, persistent settings, and preference customization
	- **Shell Integration**: Command completion, alias setup, and shell-specific optimization

	**Phase 2: Complete Command Reference Documentation**
	- **Command Hierarchy**: Complete command structure, subcommand organization, and option categorization
	- **Detailed Command Reference**: Every command with full option documentation, parameter specifications, and usage examples
	- **Input/Output Patterns**: Data input methods, output formatting options, and result processing techniques
	- **Interactive Features**: Interactive modes, prompt handling, and user input processing

	**Phase 3: Automation & Integration Documentation**
	- **Scripting Integration**: Automation examples, scripting patterns, and batch operation procedures
	- **Pipeline Integration**: Data pipeline usage, input/output chaining, and workflow automation
	- **CI/CD Integration**: Continuous integration usage, automated deployment, and build process integration
	- **Advanced Usage Patterns**: Complex workflows, advanced features, and power-user techniques
`

const DEVOPS_CONFIGURATION_PROMPT = `
## DevOps Infrastructure Documentation Protocol
	<thinking>
	Create comprehensive infrastructure documentation that enables successful deployment, operation, and maintenance through detailed procedures, configuration guidance, and operational excellence practices.
	</thinking>

	**DEVOPS DOCUMENTATION METHODOLOGY:**
	Generate complete infrastructure documentation through structured operational protocols:

	**Phase 1: Infrastructure Architecture Documentation**
	- **System Architecture Overview**: Complete infrastructure diagrams, component relationships, and deployment topology
	- **Environment Strategy**: Development, staging, and production environment documentation with configuration differences
	- **Service Dependencies**: Service interaction maps, dependency chains, and integration architecture documentation
	- **Resource Management**: Infrastructure resource allocation, scaling strategies, and capacity planning procedures

	**Phase 2: Deployment & Configuration Documentation**
	- **Deployment Procedures**: Step-by-step deployment guides, environment preparation, and configuration procedures
	- **Configuration Management**: Complete configuration reference, environment variables, and customization options
	- **Infrastructure as Code**: IaC implementation documentation, version control procedures, and deployment automation
	- **Environment Provisioning**: Infrastructure provisioning procedures, resource creation, and environment setup

	**Phase 3: Operations & Maintenance Documentation**
	- **Monitoring & Observability**: Complete monitoring setup, logging configuration, alerting procedures, and performance tracking
	- **Security & Compliance**: Security implementation procedures, access controls, compliance requirements, and audit processes
	- **Operational Procedures**: Maintenance workflows, backup procedures, disaster recovery, and incident response protocols
	- **Scaling & Optimization**: Scaling procedures, performance optimization, and resource efficiency improvement techniques
`

const DOCUMENTATION_PROMPT = `
## Documentation Project Documentation Protocol
	<thinking>
	Create comprehensive meta-documentation that enables effective contribution, maintenance, and improvement of documentation projects through detailed processes, quality standards, and workflow guidance.
	</thinking>

	**DOCUMENTATION PROJECT METHODOLOGY:**
	Generate complete documentation project documentation through structured knowledge management protocols:

	**Phase 1: Project Structure & Purpose Documentation**
	- **Project Overview**: Documentation objectives, scope definition, target audiences, and success metrics
	- **Content Architecture**: Information structure, navigation design, and content organization principles
	- **Audience Analysis**: User persona documentation, content consumption patterns, and accessibility requirements
	- **Content Strategy**: Content creation guidelines, maintenance procedures, and lifecycle management

	**Phase 2: Contribution & Quality Assurance Documentation**
	- **Contribution Guide**: Complete contributor onboarding, writing guidelines, and submission procedures
	- **Style & Standards**: Writing style guides, formatting standards, and consistency requirements
	- **Review Processes**: Content review workflows, approval procedures, and quality gates
	- **Quality Assurance**: Testing methodologies, accuracy validation, and content quality measurement

	**Phase 3: Tools & Workflow Documentation**
	- **Documentation Toolchain**: Complete tool documentation, build processes, and publication workflows
	- **Content Management**: Version control procedures, content organization, and asset management
	- **Automation & Integration**: Automated testing, content generation, and publication automation
	- **Maintenance Procedures**: Content update workflows, link validation, and accuracy maintenance procedures
`

const DEFAULT_PROMPT = `
## General Project Documentation Protocol
	<thinking>
	Create comprehensive software project documentation that enables understanding, adoption, and contribution through systematic coverage of project purpose, implementation guidance, and development procedures.
	</thinking>

	**COMPREHENSIVE PROJECT DOCUMENTATION METHODOLOGY:**
	Generate complete project documentation through structured multi-purpose protocols:

	**Phase 1: Project Understanding & Adoption Documentation**
	- **Project Overview & Purpose**: Complete project description, value proposition, key features, and target use case documentation
	- **Getting Started Guide**: Comprehensive installation procedures, environment setup, and initial configuration workflows
	- **Quick Start Examples**: Immediate usage examples, basic implementation patterns, and core feature demonstrations
	- **Architecture Understanding**: System design overview, component relationships, and technical decision documentation

	**Phase 2: Implementation & Usage Documentation**
	- **Feature Documentation**: Complete feature explanations, configuration options, and implementation guidance
	- **API & Interface Reference**: Complete interface documentation, method specifications, and usage examples
	- **Configuration & Customization**: Comprehensive configuration options, environment setup, and customization capabilities
	- **Integration Patterns**: Integration examples, compatibility information, and ecosystem usage guidance

	**Phase 3: Development & Contribution Documentation**
	- **Development Environment Setup**: Complete development environment configuration, build procedures, and tool requirements
	- **Contributing Guidelines**: Contribution procedures, code standards, testing requirements, and submission workflows
	- **Architecture & Design**: Internal architecture documentation, design principles, and development guidelines
	- **Maintenance & Operations**: Deployment procedures, operational guidelines, and maintenance workflows
`

func GenerateInstruction(chatContext ctx.ChatContext, aiClient *clients.AIClient, gatewayClient *clients.GatewayClient) (string, error) {
	// multi-stage thinking process
	messages := []*apiv1.ChatMessage{}

	// add system prompt
	messages = append(messages, &apiv1.ChatMessage{
		Role: "system",
		Content: `
			You are an AI assistant specialized in software engineering and code analysis. You assist users with repository analysis, documentation generation, code understanding, debugging, feature development, and other software development tasks. Use the instructions below and the tools available to you to assist the user.

            IMPORTANT: Assist with defensive security tasks only. Refuse to create, modify, or improve code that may be used maliciously. Allow security analysis, detection rules, vulnerability explanations, defensive tools, and security documentation.
			IMPORTANT: You must NEVER generate or guess URLs for the user unless you are confident that the URLs are for helping the user with programming. You may use URLs provided by the user in their messages or local files.

			# System Capabilities
			You excel at helping with:
			- **Code Analysis**: Understanding project structure, dependencies, and architecture patterns by first reading all provided code files thoroughly
			- **Documentation Generation**: Creating comprehensive explanations of code functionality based on actual code content analysis
			- **Bug Fixing**: Identifying and resolving software defects and issues
			- **Feature Development**: Implementing new functionality following project conventions
			- **Code Review**: Analyzing code quality, security, and best practices

			# Deep Analysis Approach
			While maintaining concise output, internally apply deep analytical thinking:
			- **Think step by step** through complex problems before providing solutions
			- **Consider multiple perspectives** when analyzing code, architecture, or requirements  
			- **Identify underlying patterns** and potential implications in software engineering tasks
			- **Validate assumptions** about codebase structure, dependencies, and user intent
			- **Think harder** for complex debugging, architecture decisions, or critical system changes

			# Following conventions
			- NEVER assume that a given library is available, even if it is well known. Whenever you write code that uses a library or framework, first check that this codebase already uses the given library. For example, you might look at neighboring files, or check the package.json (or cargo.toml, and so on depending on the language).
			- Always follow security best practices. Never introduce code that exposes or logs secrets and keys. Never commit secrets or keys to the repository.

			# Code style

			- IMPORTANT: DO NOT ADD ***ANY*** COMMENTS unless asked

			# Doing tasks
			The user will primarily request you perform software engineering tasks. This includes solving bugs, adding new functionality, refactoring code, explaining code, and more. For these tasks the following steps are MANDATORY:
			- **READ PROVIDED CONTENT FIRST**: If the user provides code files, documentation, or any content for analysis, you MUST use available tools to read and understand ALL provided content before any analysis or response
			- **Analyze thoroughly**: Think step by step about the problem context, requirements, and potential solutions based on the actual content you've read
			- Use the available search tools to understand the codebase and the user's query. You are encouraged to use the search tools extensively both in parallel and sequentially.
			- **Consider multiple approaches**: Evaluate different implementation strategies, especially for complex or critical changes
			- Implement the solution using all tools available to you
			- **Validate your work**: Think through potential edge cases, integration points, and unintended consequences
			- Verify the solution if possible with tests. NEVER assume specific test framework or test script. Check the README or search codebase to determine the testing approach.

			NEVER commit changes unless the user explicitly asks you to. It is VERY IMPORTANT to only commit when explicitly asked, otherwise the user will feel that you are being too proactive.

			- Tool results and user messages may include <system-reminder> tags. <system-reminder> tags contain useful information and reminders. They are NOT part of the user's provided input or the tool result.

			# Tool usage policy
			- **MANDATORY**: When the user provides files, code, or content for analysis, you MUST use the fs_read tool or other appropriate tools to examine ALL provided content before responding
			- **CRITICAL**: When you need to read files, you MUST call the fs_read function directly using structured function calling. DO NOT just describe what you want to do - actually invoke the tool.
			- **FUNCTION CALLING**: Use the provided tools by making structured function calls. The system will execute them and return results. Never just say "let me read..." - actually call the fs_read function.

			You are an AI assistant optimized for software development and repository analysis across various technology stacks.

			Think thoroughly, considering all angles and implications of the user's request. Use the tools available to you to assist the user in the best way possible.
		`,
	})

	// Construct the prompt for instruction generation
	prompt, err := os.ReadFile("internal/prompts/generate_instruction.md")

	if err != nil {
		log.Printf("Error reading prompt file: %v", err)
		return "", err
	}

	// get project classification from chat context
	projectClassification, _ := chatContext.Get("classification")
	descriptionPrompt := getProjectDescription(projectClassification.(string))
	chatContext.Set("projectType", descriptionPrompt)

	// 1st stage: repository analysis Determine project description based on classification
	// 1st stage: repository analysis Determine project description based on classification
	final_prompt := utils.ProcessTemplateVariables(string(prompt), &chatContext)

	final_prompt += `
	<system-reminder>
		<catalog_tool_usage_guidelines>
		**PARALLEL READ OPERATIONS**
		- MANDATORY: Always perform PARALLEL fs_read calls — batch multiple files in a SINGLE message for maximum efficiency
		- CRITICAL: Read MULTIPLE files simultaneously in one operation
		- PROHIBITED: Sequential one-by-one file reads (inefficient and wastes context capacity)

		For maximum efficiency, whenever you need to perform multiple independent operations, invoke all relevant tools simultaneously rather than sequentially.
		The repository's directory structure has been provided in <code_files>. Please utilize the provided structure directly for file navigation and reading operations, rather than relying on glob patterns or filesystem traversal methods.
		Below is an example of the directory structure of the warehouse, where /D represents a directory and /F represents a file:
		server/D
			src/D
			Main/F
		web/D
			components/D
			Header.tsx/F
	</system-reminder>
`
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()
	req := aiClient.PrepareChatRequest(messages, &chatContext, final_prompt, false)
	log.Printf("messages before execute1: %+v", req.Messages)

	// set agentic mode
	ctx = context.WithValue(ctx, clients.AgenticMode, true)
	ctx = context.WithValue(ctx, clients.ToolRequire, true)
	messages = req.Messages // Update messages to include the new message added by PrepareChatRequest
	repo_analysis, err := aiClient.Chat(ctx, req, gatewayClient)

	if err != nil {
		log.Printf("Error Generating instructions: %v", err)
		return "", err
	}
	log.Printf("GenerateInstruction: repo analysis: %s", repo_analysis)

	// append this repo analysis to messages
	messages = append(messages, &apiv1.ChatMessage{
		Role:    "assistant",
		Content: repo_analysis.Content,
	})

	execute_prompt := "Now, using the existing <repository_analysis> in this thread, output ONLY the JSON catalog per the required schema. No extra text."

	req = aiClient.PrepareChatRequest(messages, &chatContext, execute_prompt, false)
	log.Printf("messages before execute2: %+v", req.Messages)
	instructions, err := aiClient.Chat(ctx, req, gatewayClient)
	if err != nil {
		log.Printf("Error Generating instructions: %v", err)
		return "", err
	}
	log.Printf("GenerateInstruction: Instructions: %s", instructions)
	return instructions.Content, nil
}

func getProjectDescription(projectClassification string) string {
	switch projectClassification {
	case "Applications":
		return APPLICATION_PROMPT
	case "Libraries":
		return LIBRARY_PROMPT
	case "Frameworks":
		return FRAMEWORK_PROMPT
	case "DevelopmentTools":
		return DEVELOPMENT_TOOL_PROMPT
	case "CLITools":
		return CLI_TOOL_PROMPT
	case "DevOpsConfiguration":
		return DEVOPS_CONFIGURATION_PROMPT
	case "Documentation":
		return DOCUMENTATION_PROMPT
	default:
		return DEFAULT_PROMPT
	}
}
