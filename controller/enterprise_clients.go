package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type EnterpriseClient struct {
	Name        string `json:"name"`
	Industry    string `json:"industry"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Users       string `json:"users"`
}

var enterpriseClients = []EnterpriseClient{
	{Name: "TechCorp Global", Industry: "Technology", Logo: "TC", Description: "Fortune 500 tech company using QuantumClaw for their internal AI platform serving 10,000+ employees.", Users: "12K+"},
	{Name: "MediAI Health", Industry: "Healthcare", Logo: "MH", Description: "Leading healthcare AI startup processing 5M+ patient records monthly through HIPAA-compliant API gateway.", Users: "5M+"},
	{Name: "FinLogic", Industry: "Finance", Logo: "FL", Description: "Top-tier financial services firm using multi-model AI for real-time risk assessment and trading analysis.", Users: "3K+"},
	{Name: "EduSpark", Industry: "Education", Logo: "ES", Description: "Global edtech platform powering personalized learning for 2M+ students across 30 countries.", Users: "2M+"},
	{Name: "RetailAI", Industry: "E-commerce", Logo: "RA", Description: "Major e-commerce platform using AI for product recommendations, inventory forecasting, and customer service.", Users: "50M+"},
	{Name: "CloudNine Games", Industry: "Gaming", Logo: "CG", Description: "AAA game studio using generative AI for NPC dialogue, asset creation, and player behavior analysis.", Users: "10M+"},
	{Name: "LogiTech Solutions", Industry: "Logistics", Logo: "LS", Description: "Supply chain optimization platform using AI models for route planning, demand forecasting, and automation.", Users: "8K+"},
	{Name: "DataStream Analytics", Industry: "Data", Logo: "DA", Description: "Enterprise data analytics platform that integrated QuantumClaw API for natural language querying of datasets.", Users: "50K+"},
}

func ListEnterpriseClients(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    enterpriseClients,
	})
}
