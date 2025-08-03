package routes

import (
	"net/http"

	"golang_starter_kit_2025/app/controllers"
	"golang_starter_kit_2025/app/middleware"
	"golang_starter_kit_2025/app/services"
	"golang_starter_kit_2025/facades"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(route *gin.Engine) {
	// Routes untuk test table (PostgreSQL, multi koneksi, tanpa auth)
	testService := services.TestService{}
	testController := controllers.NewTestController(testService)
	testRoutes := route.Group("/tests")
	{
		testRoutes.GET("", testController.List)         // List all test
		testRoutes.GET(":id", testController.Get)       // Get test by ID
		testRoutes.POST("", testController.Create)      // Create test
		testRoutes.PUT(":id", testController.Update)    // Update test
		testRoutes.DELETE(":id", testController.Delete) // Delete test
	}
	// Apply middleware logging untuk semua route
	// route.Use(middleware.LoggerMiddleware())

	// Public route: Hello World
	controller := controllers.Controller{}
	route.GET("", controller.HelloWorld)

	// Public route: Login and Logout (no auth required)
	authService := services.AuthService{}
	authController := controllers.NewAuthController(authService)
	route.PUT("/auth/login", authController.Login)
	authRoutes := route.Group("/auth").Use(middleware.AuthMiddleware())
	{
		authRoutes.GET("/logout", authController.Logout)
		authRoutes.GET("/refresh", authController.Refresh)
	}

	// Routes untuk users (protected by AuthMiddleware)
	userService := services.UserService{}
	userController := controllers.NewUserController(userService)
	userRoutes := route.Group("/users", middleware.AuthMiddleware()) // Protect user routes
	{
		userRoutes.GET("", userController.List)
		userRoutes.GET("/:id", userController.Get)
		userRoutes.PUT("", userController.Put)
		userRoutes.DELETE("/:id", userController.Delete)
		userRoutes.POST("/:id/roles", userController.AssignRoles)
		userRoutes.GET("/:id/roles", userController.GetRoles)
	}

	// Routes untuk roles (protected by AuthMiddleware)
	roleService := services.RoleService{}
	roleController := controllers.NewRoleController(roleService)
	roleRoutes := route.Group("/roles", middleware.AuthMiddleware()) // Protect role routes
	{
		roleRoutes.GET("", roleController.List)                               // List roles
		roleRoutes.PUT("", roleController.Put)                                // Create/Update role
		roleRoutes.DELETE("/:id", roleController.Delete)                      // Delete role by ID
		roleRoutes.POST("/:id/permissions", roleController.AssignPermissions) // Assign permissions to role
		roleRoutes.GET("/:id/permissions", roleController.GetPermissions)     // Get permissions for role
	}

	// Routes untuk permissions (protected by AuthMiddleware)
	permissionService := services.PermissionService{}
	permissionController := controllers.NewPermissionController(permissionService)
	permissionRoutes := route.Group("/permissions", middleware.AuthMiddleware()) // Protect permission routes
	{
		permissionRoutes.GET("", permissionController.List)          // List all permissions
		permissionRoutes.PUT("", permissionController.Put)           // Create/Update permission
		permissionRoutes.DELETE("/:id", permissionController.Delete) // Delete permission by ID
	}

	fileController := controllers.NewFileController()
	fileRoutes := route.Group("/file")
	{
		fileRoutes.GET("/:key/:filename", fileController.ServeFile)
	}

	// KYC Routes (protected by AuthMiddleware)
	kycService := services.NewKycService()
	kycController := controllers.NewKycController(kycService)
	kycRoutes := route.Group("/kyc", middleware.AuthMiddleware())
	{
		// Base64 upload endpoints
		kycRoutes.POST("/upload-id-card", kycController.UploadIdCard) // Upload ID Card (base64)
		kycRoutes.POST("/upload-selfie", kycController.UploadSelfie)  // Upload Selfie (base64)

		// File upload endpoints
		kycRoutes.POST("/upload-id-card-file", kycController.UploadIdCardFile) // Upload ID Card (file)
		kycRoutes.POST("/upload-selfie-file", kycController.UploadSelfieFile)  // Upload Selfie (file)

		// Status and processing
		kycRoutes.GET("/status/:reference", kycController.GetStatus)      // Get KYC Status
		kycRoutes.POST("/process/:id", kycController.ProcessVerification) // Manual Process
	}

	// Database management routes (protected by AuthMiddleware)
	databaseController := controllers.NewDatabaseController()
	databaseRoutes := route.Group("/api/database")
	{
		databaseRoutes.GET("/status", databaseController.GetConnectionStatus)
		databaseRoutes.GET("/health", databaseController.HealthCheck)
		databaseRoutes.GET("/test", databaseController.TestConnection)
	}

	// Endpoint untuk mengecek kesehatan koneksi facades
	route.GET("/health", func(c *gin.Context) {
		sqlDB, err := facades.DB.DB() // Mengambil facades/sql *DB dari GORM *DB
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to get facades connection",
				"error":   err.Error(),
			})
			return
		}

		err = sqlDB.Ping() // Menggunakan sqlDB untuk ping ke facades
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "facades connection failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "facades is connected",
			"facades": "supply_chain_retail", // Sesuaikan dengan nama facades Anda
		})
	})

	// Multi-database health check endpoint (public)
	route.GET("/health/databases", func(c *gin.Context) {
		manager := facades.GetManager()
		health := make(map[string]interface{})
		connections := []string{"mysql", "postgres", "mysql_secondary"}

		allHealthy := true
		for _, connName := range connections {
			if manager.IsConnected(connName) {
				stats, err := manager.GetConnectionStats(connName)
				if err == nil {
					health[connName] = gin.H{
						"status": "healthy",
						"stats":  stats,
					}
				} else {
					health[connName] = gin.H{
						"status": "unhealthy",
						"error":  err.Error(),
					}
					allHealthy = false
				}
			} else {
				health[connName] = gin.H{
					"status": "disconnected",
				}
				allHealthy = false
			}
		}

		statusCode := http.StatusOK
		if !allHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"overall_health": allHealthy,
			"connections":    health,
		})
	})

	// Test OCR endpoint (public, no auth required)
	route.GET("/test/ocr", kycController.TestOCR)
}
