<?php
// 单个属性（现有功能）
#[Route("/api/users")]
function getUsers() {}

// 属性组 - 一个 #[] 中多个属性
#[Route("/api/users"), Method("GET")]
function getUsersWithMethod() {}

// 多个属性组合
#[Route("/api/users")]
#[Method("POST")]
#[Validate("required")]
function createUser() {}

// 带参数的复杂属性
#[Route("/api/users/{id}", methods: ["GET", "POST"], name: "user_api")]
#[Cache(ttl: 3600, tags: ["users", "api"])]
function complexAttributes() {}