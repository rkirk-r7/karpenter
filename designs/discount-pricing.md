# Discounted Pricing Support

## Overview

Karpenter is currently unaware of any discounted pricing, such as volume discounts or reserved instances/savings plans, which can lead to more expensive 
instances being chosen. This pricing is apparently only available from the "payer" account, not any other child account API's for pricing. </br>

This was made explicit recently - the price of spot instances rose sharply at the beginning of Q2 2023.
For users on default pricing this may not be noticeable, however if there is any discount for on-demand instances in an 
account, it could begin to become cheaper to use larger on-demand instances. </br>

## User Stories

* Karpenter will prioritise the cheapest node capacity type in an account based on personal modifications to EC2 pricing

## Background

[Conversation on Slack](https://kubernetes.slack.com/archives/C02SFFZSA2K/p1684246928553159)

## How Will Karpenter Handle Discounted Pricing

### Spot pricing

A multiplier can be applied to spot instance types

### On Demand pricing
